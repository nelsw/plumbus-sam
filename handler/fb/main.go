// Package provides dedicated functionality for calling the Facebook Graph API.
// This package is not to be made public through the API Gateway!
package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/model/account"
	"plumbus/pkg/util/logs"
	"strings"
	"time"
)

const (
	api             = "https://graph.facebook.com/v12.0"
	formContentType = "application/x-www-form-urlencoded"
	accountFields   = "&fields=account_id,name,account_status,created_time"
)

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
	case "account":
		return accounts()
	case "campaign":
		if v, ok := req["status"]; ok {
			return nil, updateCampaignStatus(req["id"].(string), v.(string))
		}
		fallthrough
	default:
		return nil, errors.New("bad request")
	}
}

func updateCampaignStatus(id, status string) (err error) {
	if _, err = http.Post(api+"/"+id+token()+status, formContentType, nil); err != nil {
		log.WithError(err).Error()
	}
	return
}

func accounts() (out []account.Entity, err error) {

	url := api + "/" + user() + "/adaccounts" + token() + accountFields

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

	err = json.Unmarshal(data, &out)

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

func token() string {
	return "?access_token=" + os.Getenv("tkn")
}

func user() string {
	return os.Getenv("usr")
}

func main() {
	lambda.Start(handle)
}
