package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/api"
	"plumbus/pkg/util"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	db         *dynamodb.Client
	tableName  = "plumbus_fb_revenue"
	mutex      = &sync.Mutex{}
	digitCheck = regexp.MustCompile(`^[0-9]+$`)
)

type AdAccount struct {
	ID          string  `json:"id"`
	AccountID   string  `json:"account_id"`
	Name        string  `json:"name"`
	Age         float64 `json:"age"`
	AmountSpent string  `json:"amount_spent"`
	Status      int     `json:"account_status"`
}

type Campaign struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
}

type AdSet struct {
	CampaignID   string `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`
	AccountID    string `json:"account_id"`
	AccountName  string `json:"account_name"`
	Spend        string `json:"spend"`
}

type GuardOrder int

const (
	WinningOrder GuardOrder = iota
	LosingOrder
	ExitingOrder
	MissingOrder
)

type GuardStatus string

const (
	WinningStatus = "🟢 - Winning"
	LosingStatus  = "🟡 - Losing"
	ExitingStatus = "🔴 - Exiting"
	MissingStatus = "⁉️ - Missing"
)

type CampaignGuard struct {
	Account  string      `json:"account"`
	Facebook string      `json:"facebook"`
	Sovrn    string      `json:"sovrn"`
	Spend    float64     `json:"spend"`
	Revenue  float64     `json:"revenue"`
	Profit   float64     `json:"profit"`
	Status   GuardStatus `json:"status"`
	order    GuardOrder
}

func (i CampaignGuard) toPutItemInput() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{
		TableName: aws.String("plumbus_mgr"),
		Item: map[string]types.AttributeValue{
			"account":  &types.AttributeValueMemberS{Value: i.Account},
			"facebook": &types.AttributeValueMemberS{Value: i.Facebook},
			"sovrn":    &types.AttributeValueMemberS{Value: i.Sovrn},
			"spend":    &types.AttributeValueMemberN{Value: util.FloatToDecimal(i.Spend)},
			"revenue":  &types.AttributeValueMemberN{Value: util.FloatToDecimal(i.Revenue)},
			"profit":   &types.AttributeValueMemberN{Value: util.FloatToDecimal(i.Profit)},
			"status":   &types.AttributeValueMemberS{Value: string(i.Status)},
		},
	}
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

	if request.RequestContext.HTTP.Method == http.MethodGet {
		out, err := db.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName: aws.String("plumbus_mgr"),
		})

		//PrettyPrint(out)

		if err != nil {
			return api.Err(err)
		}

		var pays []CampaignGuard

		err = attributevalue.UnmarshalListOfMaps(out.Items, &pays)
		if err != nil {
			return api.Err(err)
		}

		var body []byte
		body, err = json.Marshal(&pays)
		if err != nil {
			return api.Err(err)
		}

		return api.OK(string(body))
	}

	var rev float64
	var err error

	var results []CampaignGuard

	for _, guard := range getAllAdAccounts() {

		if rev, err = getRevenue(guard.Sovrn); err != nil || rev < 0 {
			if err != nil {
				fmt.Println(guard, err)
			}
			if rev, err = getRevenue(guard.Facebook); err != nil || rev < 0 {
				if err != nil {
					fmt.Println(guard, err)
				}
				guard.Status = MissingStatus
				guard.order = MissingOrder
				results = append(results, guard)
				continue
			}
		}

		guard.Revenue = rev
		guard.Profit = guard.Revenue - guard.Spend

		if guard.Profit > 0 {
			guard.Status = WinningStatus
			guard.order = WinningOrder
		} else if guard.Profit < -100 {
			guard.Status = ExitingStatus
			guard.order = ExitingOrder
			pauseCampaign(guard.Facebook, 0)
		} else {
			guard.Status = LosingStatus
			guard.order = LosingOrder
		}

		results = append(results, guard)
	}

	sort.SliceStable(results, func(i, j int) bool {
		ri := results[i]
		rj := results[j]
		a := int(ri.order)
		z := int(rj.order)
		if a == z {
			return ri.Profit > rj.Profit
		}
		return a < z
	})

	var out *dynamodb.ScanOutput
	if out, err = db.Scan(context.TODO(), &dynamodb.ScanInput{TableName: &tableName}); err != nil {
		return api.Err(err)
	}

	type old struct {
		Facebook string `json:"facebook"`
	}

	var olds []old

	if err = attributevalue.UnmarshalListOfMaps(out.Items, &olds); err != nil {
		return api.Err(err)
	}

	for _, o := range olds {
		dii := &dynamodb.DeleteItemInput{
			TableName: aws.String("plumbus_mgr"),
			Key:       map[string]types.AttributeValue{"facebook": &types.AttributeValueMemberS{Value: o.Facebook}},
		}
		if _, err = db.DeleteItem(context.TODO(), dii); err != nil {
			fmt.Println(err)
		}
	}

	for _, r := range results {
		if _, err = db.PutItem(context.TODO(), r.toPutItemInput()); err != nil {
			fmt.Println(r)
			fmt.Println(err)
		}
	}

	return api.OK("")
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

func getAllAdAccounts() []CampaignGuard {

	base := "https://graph.facebook.com/v12.0/10158615602243295/adaccounts"
	token := "?access_token=" + os.Getenv("tkn")
	fields := "&fields=account_id,name,age,amount_spent,account_status"
	url := base + token + fields

	var err error
	var data []AdAccount

	if data, err = getAdAccounts(url); err != nil {
		fmt.Println(err)
		panic(err)
	}

	var activeAdAccounts []AdAccount
	for _, adAccount := range data {
		if adAccount.Status == 1 || adAccount.Status == 201 {
			activeAdAccounts = append(activeAdAccounts, adAccount)
		}
	}

	actives := getActiveCampaigns(activeAdAccounts)

	var f float64
	var guards []CampaignGuard
	for accountID := range actives {

		for _, adSet := range getAllAdSetInsights(accountID) {

			if f, err = strconv.ParseFloat(adSet.Spend, 64); err != nil {
				PrettyPrint(adSet)
				fmt.Println(err)
				continue
			}

			sovrnCampaignID := adSet.CampaignID
			if spaced := strings.Split(adSet.CampaignName, " "); len(spaced) > 1 && digitCheck.MatchString(spaced[0]) {
				sovrnCampaignID = spaced[0]
			} else if scored := strings.Split(adSet.CampaignName, "_"); len(scored) > 1 && digitCheck.MatchString(scored[0]) {
				sovrnCampaignID = scored[0]
			}

			guards = append(guards, CampaignGuard{
				Account:  accountID,
				Facebook: adSet.CampaignID,
				Sovrn:    sovrnCampaignID,
				Spend:    f,
			})
		}
	}

	sort.SliceStable(guards, func(i, j int) bool {
		return guards[i].Spend > guards[j].Spend
	})

	return guards
}

func getActiveCampaigns(activeAdAccounts []AdAccount) map[string][]Campaign {
	var wg sync.WaitGroup

	actives := map[string][]Campaign{}
	for _, activeAdAccount := range activeAdAccounts {
		wg.Add(1)
		go func(activeAdAccount AdAccount) {
			defer wg.Done()
			activeCampaigns := getAllCampaignsUnderAccount(activeAdAccount.AccountID)
			if len(activeCampaigns) > 0 {
				mutex.Lock()
				actives[activeAdAccount.AccountID] = activeCampaigns
				mutex.Unlock()
			}
		}(activeAdAccount)
	}

	wg.Wait()

	return actives
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

func getAllCampaignsUnderAccount(account string) []Campaign {

	base := "https://graph.facebook.com/v12.0/act_" + account + "/campaigns"
	token := "?access_token=" + os.Getenv("tkn")
	fields := "&fields=id,account_id,name,status"
	url := base + token + fields

	var err error
	var data []Campaign

	if data, err = getCampaigns(url); err != nil {
		panic(err)
	}

	var activeCampaigns []Campaign
	for _, d := range data {
		if d.Status == "ACTIVE" {
			activeCampaigns = append(activeCampaigns, d)
		}
	}
	return activeCampaigns
}

func getCampaigns(url string) (data []Campaign, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {
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
	if next, err = getCampaigns(payload.Page.Next); err != nil {
		return
	}

	data = append(data, next...)
	return data, nil
}

func getAllAdSetInsights(adSet string) []AdSet {

	base := "https://graph.facebook.com/v12.0/act_" + adSet + "/insights"
	token := "?access_token=" + os.Getenv("tkn")
	fields := "&fields=ad_id,ad_name,adset_id,adset_name,campaign_id,campaign_name,account_id,account_name,spend"
	date := "&date_preset=today"
	level := "&level=campaign"
	url := base + token + fields + date + level

	var err error
	var data []AdSet

	if data, err = getAdSetInsights(url); err != nil {
		panic(err)
	}

	return data
}

func getAdSetInsights(url string) (data []AdSet, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}

	var payload struct {
		Data []AdSet `json:"data"`
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

	var next []AdSet
	if next, err = getAdSetInsights(payload.Page.Next); err != nil {
		return
	}

	data = append(data, next...)
	return data, nil
}

func pauseCampaign(id string, attempt int) {

	if attempt > 3 {
		fmt.Println("max attempts reached for pausing campaign ... check your code and manual pause", id)
		return
	}

	var req *http.Request
	var err error

	if req, err = http.NewRequest(http.MethodPost, "https://graph.facebook.com/v12.0/"+id, nil); err != nil {
		fmt.Println(err)
		pauseCampaign(id, attempt+1)
		return
	}

	q := req.URL.Query()
	q.Add("access_token", os.Getenv("tkn"))
	q.Add("status", "PAUSED")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 25 * time.Second}
	_, err = client.Do(req)

	if err != nil {
		fmt.Println(err)
		pauseCampaign(id, attempt+1)
		return
	}

	return
}

func Pretty(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "    ")
	return string(b)
}

func PrettyPrint(v interface{}) {
	fmt.Println(Pretty(v))
}

func main() {
	lambda.Start(handle)
}
