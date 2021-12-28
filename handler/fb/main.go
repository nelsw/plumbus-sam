package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/api"
	"regexp"
	"strings"
	"sync"
)

const (
	v12 = "https://graph.facebook.com/v12.0/"
)

var (
	db         *dynamodb.Client
	tableName  = "plumbus_fb_revenue"
	digitCheck = regexp.MustCompile(`^[0-9]+$`)
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

type Page struct {
	Next string `json:"next"`
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
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
		ForceColors:   true,
	})
	if cfg, err := config.LoadDefaultConfig(context.TODO()); err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		panic(err)
	} else {
		db = dynamodb.NewFromConfig(cfg)
	}
}

func handle(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"req": request}).Info()

	domain := request.QueryStringParameters["domain"]
	id := request.QueryStringParameters["id"]
	tkn := os.Getenv("tkn")

	switch domain {
	case "acts":
		return getAllAdAccounts()
	case "camps":
		return getAllCampaignInsights(id)
	case "ads":
		return query(v12 + id + "/ads&access_token=" + tkn) // parent id
	case "ad":
		return query(v12 + id + "&access_token=" + tkn) // entity id
	default:
		return api.Err(errors.New("unrecognized domain: " + domain))
	}
}

type AdAccount struct {
	ID          string     `json:"id"`
	AccountID   string     `json:"account_id"`
	Name        string     `json:"name"`
	Age         float64    `json:"age"`
	AmountSpent string     `json:"amount_spent"`
	Status      int        `json:"account_status"`
	Campaigns   []Campaign `json:"campaigns"`
}

func getAllAdAccounts() (events.APIGatewayV2HTTPResponse, error) {
	base := "https://graph.facebook.com/v12.0/10158615602243295/adaccounts"
	token := "?access_token=" + os.Getenv("tkn")
	fields := "&fields=account_id,name,age,amount_spent,account_status"
	url := base + token + fields

	var err error
	var data []AdAccount

	if data, err = getAdAccounts(url); err != nil {
		return api.Err(err)
	}

	var wg sync.WaitGroup
	for i, adAccount := range data {
		wg.Add(1)
		go func(i int, adAccount AdAccount) {
			data[i].Campaigns = getAllCampaigns(adAccount.AccountID)
			wg.Done()
		}(i, adAccount)
	}
	wg.Wait()

	var bytes []byte
	if bytes, err = json.Marshal(&data); err != nil {
		return api.Err(err)
	}

	return api.OK(string(bytes))
}

func getAdAccounts(url string) (data []AdAccount, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}

	var payload struct {
		Data []AdAccount `json:"data"`
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

	var next []AdAccount
	if next, err = getAdAccounts(payload.Page.Next); err != nil {
		return
	}

	data = append(data, next...)
	return data, nil
}

func getCampaignsStatus(campaignID string) string {

	base := "https://graph.facebook.com/v12.0/" + campaignID
	token := "?access_token=" + os.Getenv("tkn")
	fields := "&fields=status"
	url := base + token + fields

	var err error

	var res *http.Response
	if res, err = http.Get(url); err != nil {
		fmt.Println(err)
		return ""
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		return ""
	}

	var payload struct {
		Status string `json:"status"`
	}

	if err = json.Unmarshal(body, &payload); err != nil {
		return ""
	}

	return payload.Status
}

func getCampaignRevenue(id1, id2 string) float64 {

	var rev float64
	var err error

	if rev, err = getRevenue(id1); err != nil || rev < 0 {
		if err != nil {
			fmt.Println(id1, err)
		}
		if rev, err = getRevenue(id2); err != nil || rev < 0 {
			if err != nil {
				fmt.Println(id2, err)
			}
			return -1
		}
	}

	return rev
}

func getRevenue(id string) (float64, error) {

	out, err := db.GetItem(context.TODO(), &dynamodb.GetItemInput{
		Key:       map[string]types.AttributeValue{"campaign": &types.AttributeValueMemberS{Value: id}},
		TableName: &tableName,
	})

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		return -1, err
	}

	var r struct {
		Campaign string  `json:"campaign"`
		Revenue  float64 `json:"revenue"`
	}

	if err = attributevalue.UnmarshalMap(out.Item, &r); err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		return -1, err
	}

	if r.Campaign == "" {
		return -1, nil
	}

	return r.Revenue, nil
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
		id2 := d.CampaignID
		if spaced := strings.Split(d.CampaignName, " "); len(spaced) > 1 && digitCheck.MatchString(spaced[0]) {
			id2 = spaced[0]
		} else if scored := strings.Split(d.CampaignName, "_"); len(scored) > 1 && digitCheck.MatchString(scored[0]) {
			id2 = scored[0]
		}

		data[i].Revenue = getCampaignRevenue(id1, id2)
		data[i].Status = getCampaignsStatus(d.CampaignID)
	}

	return data
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
		if spaced := strings.Split(d.CampaignName, " "); len(spaced) > 1 && digitCheck.MatchString(spaced[0]) {
			id2 = spaced[0]
		} else if scored := strings.Split(d.CampaignName, "_"); len(scored) > 1 && digitCheck.MatchString(scored[0]) {
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
		Page `json:"paging"`
	}

	if err = json.Unmarshal(body, &payload); err != nil {
		return
	}

	if data = append(payload.Data); payload.Next == "" {
		return
	}

	var next []Data
	if next, err = get(payload.Next); err != nil {
		return
	}

	data = append(data, next...)
	return data, nil
}

func main() {
	lambda.Start(handle)
}
