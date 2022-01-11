package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/api"
	"plumbus/pkg/repo"
	"strconv"
	"sync"
)

const (
	v12 = "https://graph.facebook.com/v12.0"
)

type Payload struct {
	Data []interface{} `json:"data"`
	Page struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type Account struct {
	ID        string     `json:"id"`
	AccountID string     `json:"account_id"`
	Name      string     `json:"name"`
	Status    int        `json:"account_status"`
	Created   string     `json:"created_time"`
	Campaigns []Campaign `json:"children,omitempty"`
}

type Campaign struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	DailyBudget     string  `json:"daily_budget,omitempty"`
	RemainingBudget string  `json:"remaining_budget,omitempty"`
	Created         string  `json:"created_time"`
	Updated         string  `json:"updated_time"`
	Clicks          int     `json:"clicks"`
	Impressions     int     `json:"impressions"`
	Spend           float64 `json:"spend"`
}

func (c Campaign) insightsURL() string {
	fields := "&fields=campaign_id,clicks,impressions,spend,cpc,cpp,cpm,ctr"
	dates := "&date_preset=today"
	return v12 + "/" + c.ID + "/insights" + token() + fields + dates
}

type Insight struct {
	ID          string  `json:"campaign_id"`
	Clicks      int     `json:"clicks"`
	Impressions int     `json:"impressions"`
	Spend       float64 `json:"spend"`
	CPC         float64 `json:"cpc"` // The average cost for each click (all).
	CPP         float64 `json:"cpp"` // The average cost to reach 1,000 people. This metric is estimated.
	CPM         float64 `json:"cpm"` // The average cost for 1,000 impressions.
	CTR         float64 `json:"ctr"` // The percentage of times people saw your ad and performed a click (all).
}

func handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": request}).Info()

	if id := request.QueryStringParameters["id"]; id != "" { // account id, but not the acct_id
		return getInsightData(id)
	}

	return getMarketData()
}

func getInsightData(id string) (events.APIGatewayV2HTTPResponse, error) {
	return api.OK(id)
}

func getMarketData() (events.APIGatewayV2HTTPResponse, error) {

	var err error

	var all []interface{}
	if all, err = get(accountsUrl()); err != nil {
		return api.Err(err)
	}

	var ign map[string]interface{}
	if ign, err = accountsToIgnore(); err != nil {
		return api.Err(err)
	}

	var wg sync.WaitGroup

	var accounts []Account
	for _, a := range all {

		if len(accounts) > 10 {
			break
		}

		wg.Add(1)

		go func(a interface{}) {

			defer wg.Done()

			if _, ok := ign[a.(map[string]interface{})["account_id"].(string)]; ok {
				return
			}

			var data []byte
			if data, err = json.Marshal(&a); err != nil {
				log.WithError(err).Error()
				return
			}

			var account Account
			if err = json.Unmarshal(data, &account); err != nil {
				log.WithError(err).Error()
				return
			}

			var out []interface{}
			if out, err = get(campaignsUrl(account.ID)); err != nil {
				log.WithError(err).Error()
				return
			}

			if data, err = json.Marshal(&out); err != nil {
				log.WithError(err).Error()
				return
			}

			var campaigns []Campaign
			if err = json.Unmarshal(data, &campaigns); err != nil {
				log.WithError(err).Error()
				return
			}

			for i, c := range campaigns {

				var cc []interface{}
				if cc, err = get(c.insightsURL()); err != nil {
					log.WithError(err).Error()
					continue
				}

				if cc == nil {
					continue
				}

				m := cc[0].(map[string]interface{})
				clicksStr := m["clicks"].(string)
				impressionsStr := m["impressions"].(string)
				spendStr := m["spend"].(string)

				clicks, _ := strconv.Atoi(clicksStr)
				impressions, _ := strconv.Atoi(impressionsStr)
				spend, _ := strconv.ParseFloat(spendStr, 64)

				campaigns[i].Clicks = clicks
				campaigns[i].Impressions = impressions
				campaigns[i].Spend = spend
			}

			account.Campaigns = campaigns

			accounts = append(accounts, account)
		}(a)
	}

	wg.Wait()

	var data []byte
	if data, err = json.Marshal(&accounts); err != nil {
		return api.Err(err)
	}

	return api.OnlyOK(string(data))
}

func accountsUrl() string {
	fields := "&fields=account_id,name,account_status,created_time"
	return v12 + "/10158615602243295/adaccounts" + token() + fields
}

func campaignsUrl(ID string) string {
	fields := "&fields=id,name,status,daily_budget,budget_remaining,created_time,updated_time"
	return v12 + "/" + ID + "/campaigns" + token() + fields
}

func token() string {
	return "?access_token=" + os.Getenv("tkn")
}

func accountsToIgnore() (map[string]interface{}, error) {

	var in = dynamodb.ScanInput{TableName: ptr.String("plumbus_ignored_ad_accounts")}
	var out interface{}
	if err := repo.ScanInputAndUnmarshal(&in, &out); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	res := map[string]interface{}{}
	for _, w := range out.([]interface{}) {
		res[w.(map[string]interface{})["account_id"].(string)] = true
	}

	return res, nil
}

func get(url string) (data []interface{}, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {
		log.WithError(err).Error()
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		log.WithError(err).Error()
		return
	}

	var payload Payload
	if err = json.Unmarshal(body, &payload); err != nil {
		log.WithError(err).Error()
		return
	}

	if data = append(data, payload.Data...); payload.Page.Next == "" {
		return
	}

	var next []interface{}
	if next, err = get(payload.Page.Next); err != nil {
		log.WithError(err).Error()
		return
	}

	data = append(data, next...)
	return
}

func main() {
	lambda.Start(handle)
}
