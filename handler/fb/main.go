package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/api"
	"plumbus/pkg/repo"
	"plumbus/pkg/svc"
	"plumbus/pkg/util"
	"plumbus/pkg/util/logs"
	"strings"
)

type Data struct {
	ID string `json:"id"`

	Status int `json:"account_status,omitempty"`

	// Name of the account. If not set, the name of the first admin visible to the user will be returned.
	Name string `json:"name,omitempty"`

	// Age is the amount of time the ad account has been open, in days.
	Age float64 `json:"age,omitempty"`

	Spent string `json:"amount_spent,omitempty"`

	Campaigns struct {
		Data []struct {
			Insights struct {
				Data []struct {
					ID    string `json:"id"`
					Spend string `json:"spend"`
					Start string `json:"date_start"`
					Stop  string `json:"date_stop"`
				} `json:"data"`
			} `json:"insights"`
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"data"`
	} `json:"campaigns"`

	Picture struct {
		Data struct {
			Height int    `json:"height"`
			Width  int    `json:"width"`
			URL    string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

type Campaign struct {
	CampaignID   string  `json:"campaign_id"`
	CampaignName string  `json:"campaign_name"`
	AccountID    string  `json:"account_id"`
	AccountName  string  `json:"account_name"`
	Spend        string  `json:"spend"`
	Status       string  `json:"status"`
	Revenue      float64 `json:"revenue"`
}

func init() {
	logs.Init()
}

func handle(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"req": request}).Info()

	domain := request.QueryStringParameters["domain"]
	id := request.QueryStringParameters["id"]

	switch domain {
	case "acts":
		return accounts()
	case "camps":
		return campaigns(id)
	case "ads":
		return query("https://graph.facebook.com/v12.0/" + id + "/ads&access_token=" + os.Getenv("tkn")) // parent id
	case "ad":
		return query("https://graph.facebook.com/v12.0/" + id + "&access_token=" + os.Getenv("tkn")) // entity id
	default:
		return api.Err(errors.New("unrecognized domain: " + domain))
	}
}

func accounts() (events.APIGatewayV2HTTPResponse, error) {

	var err error

	var bytes []byte
	if bytes, err = svc.Accounts(); err != nil {
		return api.Err(err)
	}

	return api.OK(string(bytes))
}

func campaigns(id string) (events.APIGatewayV2HTTPResponse, error) {

	var err error

	var out []interface{}
	if out, err = svc.Campaigns(id); err != nil {
		return api.Err(err)
	}

	var bytes []byte
	if bytes, err = json.Marshal(&out); err != nil {
		return api.Err(err)
	}

	return api.OK(string(bytes))
}

func getCampaignsStatus(campaignID string) string {

	base := "https://graph.facebook.com/v12.0/" + campaignID
	token := "?access_token=" + os.Getenv("tkn")
	fields := "&fields=status"
	url := base + token + fields

	var err error

	var res *http.Response
	if res, err = http.Get(url); err != nil {
		log.WithError(err).Error()
		return ""
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		log.WithError(err).Error()
		return ""
	}

	var payload struct {
		Status string `json:"status"`
	}

	if err = json.Unmarshal(body, &payload); err != nil {
		log.WithError(err).Error()
		return ""
	}

	return payload.Status
}

func getCampaignRevenue(id1, id2 string) float64 {

	var rev float64
	var err error

	if rev, err = getRevenue(id1); err != nil || rev < 0 {
		if err != nil {
			log.WithError(err).Error("while getting revenue by fb camp id", id1)
		}
		if rev, err = getRevenue(id2); err != nil || rev < 0 {
			if err != nil {
				log.WithError(err).Error("while getting revenue by sovrn id", id2)
			}
			return -1
		}
	}

	return rev
}

func getRevenue(id string) (float64, error) {

	var bytes []byte
	var err error

	if bytes, err = repo.Get("plumbus_fb_revenue", "campaign", id); err != nil {
		log.WithError(err).Error()
		return -1, err
	}

	var r struct {
		Account  string  `json:"account"`
		Campaign string  `json:"campaign"`
		AdSet    string  `json:"adset"`
		Revenue  float64 `json:"revenue"`
	}

	if err = json.Unmarshal(bytes, &r); err != nil {
		log.WithError(err).Error()
		return -1, err
	}

	if rev := r.Revenue; rev == 0 && r.Campaign == "" {
		return -1, nil
	} else {
		return rev, nil
	}
}

func getAllCampaigns(accountID string) []Campaign {
	base := "https://graph.facebook.com/v12.0/act_" + accountID + "/insights"
	token := "?access_token=" + os.Getenv("tkn")
	fields := "&fields=ad_id,ad_name,adset_id,adset_name,campaign_id,campaign_name,account_id,account_name,spend"
	date := "&date_preset=today"
	level := "&level=campaign"
	url := base + token + fields + date + level

	var err error
	var data []Campaign

	if data, err = getCampaignInsights(url); err != nil {
		panic(err)
	}

	for i, d := range data {
		id1 := d.CampaignID

		sovrnCampaignID := getParenWrappedSovrnID(d.CampaignName)
		if sovrnCampaignID == "" {
			if spaced := strings.Split(d.CampaignName, " "); len(spaced) > 1 && util.IsNumber(spaced[0]) {
				sovrnCampaignID = spaced[0]
			} else if scored := strings.Split(d.CampaignName, "_"); len(scored) > 1 && util.IsNumber(scored[0]) {
				sovrnCampaignID = scored[0]
			} else {
				sovrnCampaignID = d.CampaignID
			}
		}

		data[i].Revenue = getCampaignRevenue(id1, sovrnCampaignID)
		data[i].Status = getCampaignsStatus(d.CampaignID)
	}

	return data
}

func getParenWrappedSovrnID(s string) string {
	if chunks := strings.Split(s, "("); len(chunks) > 1 {
		chunk := chunks[1]
		chunks = strings.Split(chunk, ")")
		if len(chunks) > 1 {
			return chunks[0]
		}
	}
	return ""
}

func getAllCampaignInsights(accountID string) (events.APIGatewayV2HTTPResponse, error) {

	base := "https://graph.facebook.com/v12.0/act_" + accountID + "/insights"
	token := "?access_token=" + os.Getenv("tkn")
	fields := "&fields=ad_id,ad_name,adset_id,adset_name,campaign_id,campaign_name,account_id,account_name,spend"
	date := "&date_preset=today"
	level := "&level=campaign"
	url := base + token + fields + date + level

	var err error
	var data []Campaign

	if data, err = getCampaignInsights(url); err != nil {
		panic(err)
	}

	for i, d := range data {
		id1 := d.CampaignID
		id2 := d.CampaignID
		if spaced := strings.Split(d.CampaignName, " "); len(spaced) > 1 && util.IsNumber(spaced[0]) {
			id2 = spaced[0]
		} else if scored := strings.Split(d.CampaignName, "_"); len(scored) > 1 && util.IsNumber(scored[0]) {
			id2 = scored[0]
		}

		data[i].Revenue = getCampaignRevenue(id1, id2)
		data[i].Status = getCampaignsStatus(d.CampaignID)
	}

	var bytes []byte
	if bytes, err = json.Marshal(&data); err != nil {
		return api.Err(err)
	}

	return api.OK(string(bytes))
}

func getCampaignInsights(url string) (data []Campaign, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {
		fmt.Println(err)
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}

	var payload struct {
		Data []Campaign `json:"data"`
		Page struct {
			Next string `json:"next"`
		} `json:"paging"`
	}

	if err = json.Unmarshal(body, &payload); err != nil {
		return
	}

	if data = append(payload.Data); payload.Page.Next == "" {
		return
	}

	var next []Campaign
	if next, err = getCampaignInsights(payload.Page.Next); err != nil {
		return
	}

	data = append(data, next...)
	return data, nil
}

func query(url string) (events.APIGatewayV2HTTPResponse, error) {

	var err error
	var data []Data

	if data, err = get(url); err != nil {
		return api.Err(err)
	}

	var bytes []byte
	if bytes, err = json.Marshal(&data); err != nil {
		return api.Err(err)
	}

	return api.OK(string(bytes))
}

func get(url string) (data []Data, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}

	var payload struct {
		Data []Data `json:"data"`
		Page struct {
			Next string `json:"next"`
		} `json:"paging"`
	}

	if err = json.Unmarshal(body, &payload); err != nil {
		return
	}

	if data = append(payload.Data); payload.Page.Next == "" {
		return
	}

	var next []Data
	if next, err = get(payload.Page.Next); err != nil {
		return
	}

	data = append(data, next...)
	return data, nil
}

func main() {
	lambda.Start(handle)
}
