// Package agg provides functionality for storing aggregate data.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/api"
	"plumbus/pkg/model/fb"
	"plumbus/pkg/model/sovrn"
	"plumbus/pkg/repo"
	"plumbus/pkg/util"
	"plumbus/pkg/util/logs"
	"strings"
	"sync"
	"time"
)

const (
	v12            = "https://graph.facebook.com/v12.0"
	accountFields  = "&fields=account_id,name,account_status,created_time"
	campaignFields = "&fields=account_id,id,name,status,daily_budget,budget_remaining,created_time,updated_time"
)

var (
	accountTable  = "plumbus_fb_account"
	campaignTable = "plumbus_fb_campaign"
	mutex         = sync.Mutex{}
)

type account struct {
	ID      string `json:"account_id"`
	Name    string `json:"name"`
	Status  int    `json:"account_status"`
	Created string `json:"created_time"`
}

func (a account) putRequest() (*types.PutRequest, error) {
	if item, err := attributevalue.MarshalMap(&a); err != nil {
		log.Trace(err)
		return nil, err
	} else {
		return &types.PutRequest{Item: item}, nil
	}
}

func (a account) writeRequest() (out types.WriteRequest, err error) {
	if out.PutRequest, err = a.putRequest(); err != nil {
		log.Trace(err)
	}
	return
}

func (a account) campaignsURL() string {
	return v12 + "/act_" + a.ID + "/campaigns" + token() + campaignFields
}

func insightsURL(accountID string) string {
	fields := "&fields=campaign_id,clicks,impressions,spend,cpc,cpp,cpm,ctr"
	dates := "&date_preset=today"
	level := "&level=campaign"
	return v12 + "/act_" + accountID + "/insights" + token() + fields + dates + level
}

type campaign struct {
	AccountID       string `json:"account_id"`
	ID              string `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	DailyBudget     string `json:"daily_budget"`
	RemainingBudget string `json:"remaining_budget"`
	Created         string `json:"created_time"`
	Updated         string `json:"updated_time"`

	Clicks      string `json:"clicks"`
	Impressions string `json:"impressions"`
	Spend       string `json:"spend"`
	CPC         string `json:"cpc"` // The average cost for each click (all).
	CPP         string `json:"cpp"` // The average cost to reach 1,000 people. This metric is estimated.
	CPM         string `json:"cpm"` // The average cost for 1,000 impressions.
	CTR         string `json:"ctr"` // The percentage of times people saw your ad and performed a click (all).

	UTM     string  `json:"utm"`
	Revenue float64 `json:"revenue"`
}

func (c campaign) utm() string {

	if chunks := strings.Split(c.Name, "("); len(chunks) > 1 {

		if chunks = strings.Split(chunks[1], ")"); len(chunks) > 1 {
			return chunks[0]
		}
	}

	if spaced := strings.Split(c.Name, " "); len(spaced) > 1 && util.IsNumber(spaced[0]) {
		return spaced[0]
	}

	if scored := strings.Split(c.Name, "_"); len(scored) > 1 && util.IsNumber(scored[0]) {
		return scored[0]
	}

	return c.ID
}

func (c campaign) writeRequest() (out types.WriteRequest, err error) {
	var item map[string]types.AttributeValue
	if item, err = attributevalue.MarshalMap(&c); err != nil {
		log.WithError(err).Error()
	} else {
		out.PutRequest = &types.PutRequest{Item: item}
	}
	return
}

func (c campaign) insightsURL() string {
	fields := "&fields=campaign_id,clicks,impressions,spend,cpc,cpp,cpm,ctr"
	dates := "&date_preset=today"
	return v12 + "/" + c.ID + "/insights" + token() + fields + dates
}

type Insight struct {
	CampaignID  string `json:"campaign_id"`
	Clicks      string `json:"clicks"`
	Cpc         string `json:"cpc"`
	Cpm         string `json:"cpm"`
	Cpp         string `json:"cpp"`
	Ctr         string `json:"ctr"`
	DateStart   string `json:"date_start"`
	DateStop    string `json:"date_stop"`
	Impressions string `json:"impressions"`
	Spend       string `json:"spend"`
}

func init() {
	logs.Init()
}

// get account and campaign data from fb, save it
// get account and campaign data from fb, save it
// get campaign insights, save them
// get sovrn data, save it
func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": req}).Info()

	method := req.RequestContext.HTTP.Method

	if method == http.MethodOptions {
		return api.OK("")
	}

	node := req.QueryStringParameters["node"]

	if method == http.MethodGet {

		if node == "account" {
			return getAccountEntitiesResponse()
		}

		if node == "campaign" {
			return getCampaignEntitiesResponse(ctx, req.QueryStringParameters["id"])
		}

	}

	if method == http.MethodPut {

		if node == "account" {
			return putAccountValuesResponse(ctx)
		}

		if node == "campaign" {
			return putCampaignDetailValuesResponse(ctx)
		}
	}

	return api.Err(errors.New("nothing handled"))

}

func putAccountValuesResponse(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	var err error

	var out []account
	if out, err = getAccountValues(); err != nil {
		return api.Err(err)
	}

	qty := len(out)

	var w types.WriteRequest
	var ww []types.WriteRequest
	for i, o := range out {

		if w, err = o.writeRequest(); err != nil {
			log.Trace(err)
			continue
		}

		if ww = append(ww, w); len(ww)%25 == 0 || i == qty {
			if _, err = repo.BatchWriteItems(ctx, accountTable, ww); err != nil {
				log.Trace(err)
			}
			ww = []types.WriteRequest{}
		}
	}

	data, _ := json.Marshal(&out)
	return api.OK(string(data))
}

func putCampaignDetailValuesResponse(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	var err error

	var aa []account
	if err = repo.Scan(&dynamodb.ScanInput{TableName: &accountTable}, &aa); err != nil {
		return api.Err(err)
	}

	out := map[string][]campaign{}

	var wg sync.WaitGroup
	for _, a := range aa {

		wg.Add(1)

		go func(a account) {

			defer wg.Done()

			var all []interface{}
			if all, err = get(a.campaignsURL()); err != nil {
				log.WithError(err).Error()
				return
			}

			var data []byte
			if data, err = json.Marshal(&all); err != nil {
				log.WithError(err).Error()
				return
			}

			var cc []campaign
			if err = json.Unmarshal(data, &cc); err != nil {
				log.WithError(err).Error()
				return
			}

			mutex.Lock()
			out[a.ID] = cc
			mutex.Unlock()
		}(a)
	}

	wg.Wait()

	if err = addInsightData(ctx, out); err != nil {
		return api.Err(err)
	}

	return api.OK("")
}

func addInsightData(ctx context.Context, out map[string][]campaign) (err error) {

	var wg sync.WaitGroup

	for accountID, c := range out {

		wg.Add(1)

		go func(accountID string, cc []campaign) {

			defer wg.Done()

			var all []interface{}
			if all, err = get(insightsURL(accountID)); err != nil {
				log.Trace(err)
				saveData(ctx, cc)
				return
			}

			if len(all) == 0 {
				saveData(ctx, cc)
				return
			}

			var data []byte
			if data, err = json.Marshal(&all); err != nil {
				log.Trace(err)
				saveData(ctx, cc)
				return
			}

			var ii []Insight
			if err = json.Unmarshal(data, &ii); err != nil {
				log.Trace(err)
				saveData(ctx, cc)
				return
			}

			addRevenue(ctx, cc, ii)

		}(accountID, c)
	}

	wg.Wait()

	return
}

func saveData(ctx context.Context, cc []campaign) {

	var rr []types.WriteRequest
	for _, c := range cc {
		if r, err := c.writeRequest(); err != nil {
			log.Trace(err)
		} else {
			rr = append(rr, r)
		}
	}

	for _, c := range chunkSlice(rr, 25) {
		if _, err := repo.BatchWriteItems(ctx, campaignTable, c); err != nil {
			log.WithError(err).Error()
		}
	}
}

func addRevenue(ctx context.Context, cc []campaign, ii []Insight) {

	m := map[string]campaign{}
	for _, c := range cc {
		m[c.ID] = c
	}

	var wg sync.WaitGroup

	var r types.WriteRequest
	var rr []types.WriteRequest
	for _, i := range ii {

		wg.Add(1)

		go func(i Insight) {

			defer wg.Done()

			c := m[i.CampaignID]
			c.Clicks = i.Clicks
			c.CPM = i.Cpm
			c.CTR = i.Ctr
			c.CPP = i.Cpp
			c.CPC = i.Cpc
			c.Spend = i.Spend
			c.Impressions = i.Impressions

			var err error

			var v sovrn.Entity
			if err = repo.GetIt(v.Table(), "UTM", c.utm(), &v); err != nil {
				log.Trace(err)
			} else {
				c.Revenue = v.Revenue
				c.UTM = v.UTM
			}

			if r, err = c.writeRequest(); err != nil {
				log.Trace(err)
			} else {
				rr = append(rr, r)
			}
		}(i)
	}

	wg.Wait()

	for _, c := range chunkSlice(rr, 25) {
		if _, err := repo.BatchWriteItems(ctx, campaignTable, c); err != nil {
			log.WithError(err).Error()
		}
	}
}

func chunkSlice(slice []types.WriteRequest, size int) (chunks [][]types.WriteRequest) {
	var end int
	for i := 0; i < len(slice); i += size {
		if end = i + size; end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return
}

func getAccountEntitiesResponse() (events.APIGatewayV2HTTPResponse, error) {

	var err error

	var out []account
	if err = repo.Scan(&dynamodb.ScanInput{TableName: &accountTable}, &out); err != nil {
		return api.Err(err)
	}

	var data []byte
	if data, err = json.Marshal(&out); err != nil {
		return api.Err(err)
	}

	return api.OK(string(data))
}

func getAccountValues() (out []account, err error) {

	var all []interface{}
	if all, err = get(accountsUrl()); err != nil {
		return nil, err
	}

	var ign map[string]interface{}
	if ign, err = accountsToIgnore(); err != nil {
		return
	}

	var wg sync.WaitGroup

	for _, v := range all {

		wg.Add(1)

		go func(v interface{}) {

			defer wg.Done()

			if _, ok := ign[v.(map[string]interface{})["account_id"].(string)]; ok {
				return
			}

			var data []byte
			if data, err = json.Marshal(&v); err != nil {
				log.WithError(err).Error()
				return
			}

			var a account
			if err = json.Unmarshal(data, &a); err != nil {
				log.WithError(err).Error()
				return
			}

			if a == (account{}) {
				log.Trace("empty account ", string(data))
				return
			}

			out = append(out, a)
		}(v)
	}

	wg.Wait()

	return
}

func accountsUrl() string {
	return v12 + "/" + user() + "/adaccounts" + token() + accountFields
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

func getCampaignEntitiesResponse(ctx context.Context, accountID string) (events.APIGatewayV2HTTPResponse, error) {

	var err error

	in := &dynamodb.QueryInput{
		TableName:              &campaignTable,
		KeyConditionExpression: ptr.String("AccountID = :v1"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberS{Value: accountID},
		},
	}

	var out *dynamodb.QueryOutput
	if out, err = repo.Query(ctx, in); err != nil {
		return api.Err(err)
	}

	var camps []campaign
	if err = attributevalue.UnmarshalListOfMaps(out.Items, &camps); err != nil {
		return api.Err(err)
	}

	if len(camps) == 0 {
		return api.Empty()
	}

	var data []byte
	if data, err = json.Marshal(&camps); err != nil {
		return api.Err(err)
	}

	return api.OK(string(data))
}

/*
	fb base
*/
func get(url string) (data []interface{}, err error) {
	return getAttempt(url, 1)
}

func getAttempt(url string, attempt int) (data []interface{}, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {

		str := err.Error()

		if !strings.Contains(str, "too many open files") &&
			!strings.Contains(str, "no such host") {
			log.WithError(err).Error()
		}

		if strings.Contains(str, "connection refused") {
			log.Trace("chill out")
			fmt.Println(err)
			fmt.Println(err.Error())
			time.Sleep(time.Second * time.Duration(attempt))
			return getAttempt(url, attempt+1)
		}

		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		log.WithError(err).Error()
		return
	}

	var payload fb.Payload
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

func token() string {
	return "?access_token=" + os.Getenv("tkn")
}

func user() string {
	return os.Getenv("usr")
}

func main() {
	lambda.Start(handle)
}
