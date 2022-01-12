// Package agg provides functionality for storing aggregate data.
// Processes multiple data sources in a "today" time range.
// When we begin looking at data sources by hour, use this code:
/*
	n := time.Now().Local()
	dateYesterday := n.Add(time.Hour * 24 * -1).Format(dateLayout)
	dateToday := n.Format(dateLayout)
	dateTomorrow := n.Add(time.Hour * 24).Format(dateLayout)
	b1 := "{since:'" + dateYesterday + "',until:'" + dateToday + "'}"
	b2 := "{since:'" + dateToday + "',until:'" + dateTomorrow + "'}"
	a.url += "&time_ranges=[" + strings.Join([]string{b1, b2}, ",") + "]"
*/
package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/model/fb"
	"plumbus/pkg/model/sovrn"
	"plumbus/pkg/repo"
	"plumbus/pkg/sam"
	"plumbus/pkg/util"
	"plumbus/pkg/util/logs"
	"strconv"
	"strings"
	"sync"
)

const (
	handlerName    = "plumbus_aggHandler"
	accountFields  = "&fields=account_id,name,account_status,created_time"
	campaignFields = "&fields=account_id,id,name,status,daily_budget,budget_remaining,created_time,updated_time"
	insightsFields = "&fields=campaign_id,clicks,impressions,spend,cpc,cpp,cpm,ctr"
)

var (
	accountTable  = "plumbus_fb_account"
	campaignTable = "plumbus_fb_campaign"
	mutex         = &sync.Mutex{}
)

type accountNode struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Status string         `json:"status"`
	Kids   []campaignNode `json:"children"`
}

type account struct {
	ID      string `json:"account_id"`
	Name    string `json:"name"`
	Status  int    `json:"account_status"`
	Created string `json:"created_time"`
}

func (a account) writeRequest() (out types.WriteRequest, err error) {
	var item map[string]types.AttributeValue
	if item, err = attributevalue.MarshalMap(&a); err == nil {
		out.PutRequest = &types.PutRequest{Item: item}
	}
	return
}

func (a account) campaignsURL() string {
	return fb.API() + "/act_" + a.ID + "/campaigns" + fb.Token() + campaignFields
}

func insightsURL(ID string) string {
	return fb.API() + "/act_" + ID + "/insights" + fb.Token() + insightsFields + "&level=campaign&date_preset=today"
}

type campaignNode struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
}

type campaign struct {
	AccountID       string `json:"account_id"`
	ID              string `json:"id"`
	Name            string `json:"name"`
	Circ            string `json:"status"`
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
	Profit  float64 `json:"profit"`
	ROI     float64 `json:"roi"`
}

func (c campaign) spend() (out float64) {
	if c.Spend != "" {
		var err error
		if out, err = strconv.ParseFloat(c.Spend, 64); err != nil {
			log.WithError(err).Trace("spend ", c.Spend)
		}
	}
	return
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

type insight struct {
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

	node := req.QueryStringParameters["node"]

	switch req.RequestContext.HTTP.Method {

	case http.MethodOptions:
		return api.K()

	case http.MethodGet:
		switch node {
		case "root":
			return getRoot(ctx)
		case "account":
			return getAccounts(ctx)
		case "campaign":
			return getCampaigns(ctx, req.QueryStringParameters["id"])
		default:
			return api.Nada()
		}

	case http.MethodPut:
		switch node {
		case "account":
			return putAccountValuesResponse(ctx)
		case "campaign":
			return putCampaignDetailValuesResponse(ctx)
		default:
			return api.Nada()
		}

	case http.MethodPost:
		switch node {
		case "root":
			return postRoot(ctx)
		case "account":
			return postAccounts(ctx)
		case "campaign":
			return postCampaigns(ctx)
		default:
			return api.Nada()
		}

	default:
		return api.Nada()
	}
}

func postRoot(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	data := api.NewRequestBytes(http.MethodPost, map[string]string{"node": "account"})
	if _, err := sam.NewEvent(ctx, handlerName, data); err != nil {
		return api.Err(err)
	}
	return api.K()
}

func postAccounts(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	if res, _ := putAccountValuesResponse(ctx); res.StatusCode != http.StatusOK {
		return res, nil
	}
	data := api.NewRequestBytes(http.MethodPost, map[string]string{"node": "campaign"})
	if _, err := sam.NewEvent(ctx, handlerName, data); err != nil {
		return api.Err(err)
	}
	return api.K()
}

func postCampaigns(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	if res, _ := putCampaignDetailValuesResponse(ctx); res.StatusCode != http.StatusOK {
		return res, nil
	}
	data := api.NewRequestBytes(http.MethodPost, nil)
	if _, err := sam.NewEvent(ctx, "plumbus_ruleHandler", data); err != nil {
		return api.Err(err)
	}
	return api.K()
}

func putAccountValuesResponse(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	if aa, err := getAccountValues(ctx); err != nil {
		return api.Err(err)
	} else {

		var ww []types.WriteRequest

		for _, a := range aa {
			if w, err := a.writeRequest(); err != nil {
				log.WithFields(log.Fields{"id": a.ID}).
					WithError(err).
					Error("converting an account value to a write request")
			} else {
				ww = append(ww, w)
			}
		}

		if err := repo.BatchWriteItems(ctx, accountTable, ww); err != nil {
			log.WithError(err).Error("agg account value batch write items")
			return api.Err(err)
		}

		return api.K()
	}
}

func putCampaignDetailValuesResponse(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	var err error

	var aa []account
	if err = repo.Scan(ctx, &dynamodb.ScanInput{TableName: &accountTable}, &aa); err != nil {
		return api.Err(err)
	}

	out := map[string][]campaign{}

	var wg sync.WaitGroup
	for _, a := range aa {

		wg.Add(1)

		go func(a account) {

			defer wg.Done()

			var all []interface{}
			if all, err = fb.Get(a.campaignsURL()); err != nil {
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
			if ko, ok := out[a.ID]; ok {
				cc = append(cc, ko...)
			}
			out[a.ID] = cc
			mutex.Unlock()
		}(a)
	}

	wg.Wait()

	addInsightData(ctx, out)

	return api.OK("")
}

func addInsightData(ctx context.Context, out map[string][]campaign) {

	var wg sync.WaitGroup

	for accountID, c := range out {

		wg.Add(1)

		go func(accountID string, cc []campaign) {

			defer wg.Done()

			var err error

			var all []interface{}
			if all, err = fb.Get(insightsURL(accountID)); err != nil {
				log.WithError(err).Error()
				saveCampaigns(ctx, cc)
				return
			}

			if len(all) == 0 {
				saveCampaigns(ctx, cc)
				return
			}

			var data []byte
			if data, err = json.Marshal(&all); err != nil {
				log.WithError(err).Error()
				saveCampaigns(ctx, cc)
				return
			}

			var ii []insight
			if err = json.Unmarshal(data, &ii); err != nil {
				log.WithError(err).Error()
				saveCampaigns(ctx, cc)
				return
			}

			addRevenue(ctx, cc, ii)

		}(accountID, c)
	}

	wg.Wait()

	return
}

func saveCampaigns(ctx context.Context, cc []campaign) {

	var rr []types.WriteRequest
	for _, c := range cc {
		if r, err := c.writeRequest(); err != nil {
			log.WithError(err).Error()
		} else {
			rr = append(rr, r)
		}
	}

	if err := repo.BatchWriteItems(ctx, campaignTable, rr); err != nil {
		log.WithError(err).Error()
	}
}

func batchGet(ctx context.Context, keys []map[string]types.AttributeValue) (got []sovrn.Entity, err error) {
	in := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			"plumbus_fb_sovrn": {
				Keys: keys,
			},
		},
	}
	var out *dynamodb.BatchGetItemOutput
	if out, err = repo.BatchGetItem(ctx, in); err != nil {
		log.WithError(err).Error()
	} else if err = attributevalue.UnmarshalListOfMaps(out.Responses["plumbus_fb_sovrn"], &got); err != nil {
		log.WithError(err).Error()
	}
	return
}

func addRevenue(ctx context.Context, cc []campaign, ii []insight) {

	iii := map[string]insight{}
	for _, i := range ii {
		iii[i.CampaignID] = i
	}

	qqq := map[string]campaign{}
	sss := map[string]sovrn.Entity{}
	var keys []map[string]types.AttributeValue
	for idx, c := range cc {

		if _, ok := qqq[c.utm()]; ok {
			continue
		}

		keys = append(keys, map[string]types.AttributeValue{"UTM": &types.AttributeValueMemberS{Value: c.utm()}})
		qqq[c.utm()] = c

		if len(keys)%25 == 0 || idx == len(cc) {
			if ss, err := batchGet(ctx, keys); err != nil {
				log.WithError(err).Error()
			} else {
				for _, s := range ss {
					sss[s.UTM] = s
				}
			}
			keys = []map[string]types.AttributeValue{}
		}
	}

	var wg sync.WaitGroup

	var rr []types.WriteRequest
	for _, c := range cc {

		wg.Add(1)

		go func(c campaign) {

			defer wg.Done()

			if i, ok := iii[c.ID]; ok {
				c.Clicks = i.Clicks
				c.CPM = i.Cpm
				c.CTR = i.Ctr
				c.CPP = i.Cpp
				c.CPC = i.Cpc
				c.Spend = i.Spend
				c.Impressions = i.Impressions
			}

			if c.Spend == "" {
				c.Spend = "0"
			}

			if s, ok := sss[c.utm()]; ok {
				c.UTM = s.UTM
				c.Revenue = s.Revenue
			}

			c.Profit = c.Revenue - c.spend()

			if c.Profit != 0 && c.spend() != 0 {
				c.ROI = c.Profit / c.spend() * 100
			} else if c.Profit != 0 {
				c.ROI = c.Profit * 100
			} else if c.spend() != 0 {
				c.ROI = c.spend() * 100
			} else {
				c.ROI = 0
			}

			if r, err := c.writeRequest(); err != nil {
				log.WithError(err).Trace()
			} else {
				rr = append(rr, r)
			}
		}(c)
	}

	wg.Wait()

	if err := repo.BatchWriteItems(ctx, campaignTable, rr); err != nil {
		log.WithError(err).Error()
	}
}

func getAccounts(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	var out []account
	if err := repo.Scan(ctx, &dynamodb.ScanInput{TableName: &accountTable}, &out); err != nil {
		return api.Err(err)
	}

	if ign, err := fb.AccountsToIgnore(ctx); err != nil {
		return api.Err(err)
	} else {

		var aa []account
		for _, a := range out {
			if _, ok := ign[a.ID]; !ok {
				aa = append(aa, a)
			}
		}

		if data, err := json.Marshal(&aa); err != nil {
			return api.Err(err)
		} else {
			return api.OK(string(data))
		}
	}
}

func getAccountValues(ctx context.Context) (out []account, err error) {

	var all []interface{}
	if all, err = fb.Get(fb.API() + "/" + fb.User() + "/adaccounts" + fb.Token() + accountFields); err != nil {
		return nil, err
	}

	var ign map[string]interface{}
	if ign, err = fb.AccountsToIgnore(ctx); err != nil {
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

			out = append(out, a)
		}(v)
	}

	wg.Wait()

	return
}

func getCampaigns(ctx context.Context, accountID string) (events.APIGatewayV2HTTPResponse, error) {

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

func getRoot(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	var aa []accountNode
	if err := repo.Scan(ctx, &dynamodb.ScanInput{TableName: &accountTable}, &aa); err != nil {
		return api.Err(err)
	}

	if ign, err := fb.AccountsToIgnore(ctx); err != nil {
		return api.Err(err)
	} else {
		var out []accountNode
		var wg sync.WaitGroup
		for _, a := range aa {
			if _, ok := ign[a.ID]; ok {
				continue
			}
			wg.Add(1)
			go func(a accountNode) {
				defer wg.Done()
				var err error
				if a.Kids, err = campaigns(ctx, a.ID); err != nil {
					log.WithError(err).Error()
				}
				out = append(out, a)
			}(a)
		}
		wg.Wait()

		if data, err := json.Marshal(&out); err != nil {
			return api.Err(err)
		} else {
			return api.OnlyOK(string(data))
		}
	}
}

func campaigns(ctx context.Context, accountID string) (cc []campaignNode, err error) {

	in := &dynamodb.QueryInput{
		TableName:              &campaignTable,
		KeyConditionExpression: ptr.String("AccountID = :v1"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberS{Value: accountID},
		},
	}

	var out *dynamodb.QueryOutput
	if out, err = repo.Query(ctx, in); err != nil {
		log.WithError(err).Error()
	} else if err = attributevalue.UnmarshalListOfMaps(out.Items, &cc); err != nil {
		log.WithError(err).Error()
	}

	return
}

func main() {
	lambda.Start(handle)
}
