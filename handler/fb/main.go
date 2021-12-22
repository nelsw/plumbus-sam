package main

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	api "plumbus/pkg"
)

const (
	v12 = "https://graph.facebook.com/v12.0/"
	act = "10158615602243295/adaccounts?fields=account_status%2Ccampaigns%7Binsights%7Bspend%2Cad_id%2Cad_name%7D%2Cstatus%2Cname%7D%2Camount_spent%2Cname%2Cage%2Cpicture"
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

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
		ForceColors:   true,
	})
}

func handle(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"req": request}).Info()

	domain := request.QueryStringParameters["domain"]
	id := request.QueryStringParameters["id"]
	tkn := os.Getenv("tkn")

	switch domain {
	case "acts":
		return query(v12 + act + "&access_token=" + tkn)
	case "ads":
		return query(v12 + id + "/ads&access_token=" + tkn) // parent id
	case "ad":
		return query(v12 + id + "&access_token=" + tkn) // entity id
	default:
		return api.Err(errors.New("unrecognized domain: " + domain))
	}
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
