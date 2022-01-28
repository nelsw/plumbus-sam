// Package provides dedicated functionality for calling the Facebook Graph API.
// This package is not to be made public through the API Gateway!
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/model/account"
	"plumbus/pkg/model/campaign"
	"plumbus/pkg/util/logs"
	"strings"
	"sync"
	"time"
)

const (
	api                = "https://graph.facebook.com/v12.0"
	formContentType    = "application/x-www-form-urlencoded"
	accountFieldsParam = "fields=account_id,name,account_status,created_time"
	descFieldsParam    = "fields=account_id,id,name,status,daily_budget,budget_remaining,created_time,updated_time"
	sightFieldsParam   = "fields=account_id,campaign_id,clicks,impressions,spend,cpc,cpp,cpm,ctr"
	levelParam         = "level=campaign"
	dateParam          = "date_preset=today"
)

var mutex = sync.Mutex{}

type payload struct {
	Data []interface{} `json:"data"`
	Page struct {
		Next string `json:"next"`
	} `json:"paging"`
}

func init() {
	logs.Init()
}

func handle(ctx context.Context, req map[string]interface{}) (interface{}, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": req}).Info()

	switch req["node"] {
	case "accounts":
		return accounts()
	case "campaign":
		return postCampaignStatus(req)
	case "campaigns":
		return getCampaigns(req)
	default:
		return nil, errors.New("bad request")
	}
}

func getCampaigns(req map[string]interface{}) (out []campaign.Entity, err error) {

	ID := req["ID"].(string)

	var insights []campaign.Entity
	if insights, err = getCampaignsSight(ID); err != nil {
		log.WithError(err).Error()
		return
	}

	ccc := map[string]campaign.Entity{}
	for _, c := range insights {
		ccc[c.CampaignID] = c
	}

	var desc []campaign.Entity
	if desc, err = getCampaignDesc(ID); err != nil {
		log.WithError(err).Error()
		return
	}

	var wg sync.WaitGroup

	for _, d := range desc {

		wg.Add(1)

		go func(d campaign.Entity) {

			defer wg.Done()

			if v, ok := ccc[d.ID]; ok {
				d.Clicks = v.Clicks
				d.Impressions = v.Impressions
				d.Spend = v.Spend
				d.CPC = v.CPC
				d.CPP = v.CPP
				d.CPM = v.CPM
				d.CTR = v.CTR
			}

			d.SetUTM()
			mutex.Lock()
			out = append(out, d)
			mutex.Unlock()
		}(d)
	}

	wg.Wait()

	log.Trace("got ", len(out), " campaign aggregates for AccountID ", ID)
	return
}

func getCampaignDesc(ID string) (cc []campaign.Entity, err error) {

	url := fmt.Sprintf("%s/act_%s/campaigns?%s&%s", api, ID, tokenParam(), descFieldsParam)

	var all []interface{}
	if all, err = get(url); err != nil {
		log.WithError(err).Error()
		return
	}

	var data []byte
	if data, err = json.Marshal(&all); err != nil {
		log.WithError(err).Error()
		return
	}

	if err = json.Unmarshal(data, &cc); err != nil {
		log.WithError(err).Error()
		return
	}

	log.Trace("got ", len(cc), " campaign descriptions for AccountID ", ID)
	return
}

func getCampaignsSight(ID string) (out []campaign.Entity, err error) {

	url := fmt.Sprintf("%s/act_%s/insights?%s&%s&%s&%s", api, ID, tokenParam(), sightFieldsParam, levelParam, dateParam)

	var all []interface{}
	if all, err = get(url); err != nil {
		log.WithError(err).Error()
		return
	}

	var data []byte
	if data, err = json.Marshal(&all); err != nil {
		log.WithError(err).Error()
		return
	}

	if err = json.Unmarshal(data, &out); err != nil {
		log.WithError(err).Error()
		return
	}

	log.Trace("got ", len(out), " campaign insights for AccountID ", ID)

	return
}

func postCampaignStatus(req map[string]interface{}) (v interface{}, err error) {

	url := fmt.Sprintf("%s/%s?%s&%s", api, req["ID"], tokenParam(), req["status"].(campaign.Status).Param())

	if _, err = http.Post(url, formContentType, nil); err != nil {
		log.WithError(err).Error()
	}

	return
}

func accounts() (out []account.Entity, err error) {

	url := fmt.Sprintf("%s/%s/adaccounts?%s&%s", api, user(), tokenParam(), accountFieldsParam)

	var all []interface{}
	if all, err = get(url); err != nil {
		log.WithError(err).Error()
		return
	}

	var data []byte
	if data, err = json.Marshal(&all); err != nil {
		log.WithError(err).Error()
		return
	}

	if err = json.Unmarshal(data, &out); err != nil {
		log.WithError(err).Error()
	}

	return
}

func get(url string, attempts ...int) (data []interface{}, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {

		var attempt int
		if attempts != nil && len(attempts) > 0 {
			attempt = attempts[0]
		}

		if attempt > 9 {
			log.WithError(err).Error()
			return
		}

		if !strings.Contains(err.Error(), "too many open files") &&
			!strings.Contains(err.Error(), "no such host") {
			log.WithError(err).Trace()
		} else if strings.Contains(err.Error(), "connection refused") {
			log.WithError(err).Warn()
		}

		time.Sleep(time.Second * time.Duration(attempt))

		return get(url, attempt+1)
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		log.WithError(err).Error()
		return
	}

	var p payload
	if err = json.Unmarshal(body, &p); err != nil {
		log.WithError(err).Error()
		return
	}

	if data = append(data, p.Data...); p.Page.Next == "" {
		return
	}

	var next []interface{}
	if next, err = get(p.Page.Next); err != nil {
		log.WithError(err).Error()
		return
	}

	data = append(data, next...)
	return
}

func tokenParam() string {
	return "access_token=" + os.Getenv("tkn")
}

func user() string {
	return os.Getenv("usr")
}

func main() {
	lambda.Start(handle)
}
