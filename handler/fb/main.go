package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/api"
	"plumbus/pkg/util/logs"
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

var (
	sam *faas.Client
)

func init() {
	logs.Init()
	if cfg, err := config.LoadDefaultConfig(context.TODO()); err != nil {
		log.WithError(err).Fatal()
	} else {
		sam = faas.NewFromConfig(cfg)
	}
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
	b, _ := json.Marshal(map[string]interface{}{"accounts": true})
	input := &faas.InvokeInput{
		FunctionName:   ptr.String("plumbus_accountHandler"),
		InvocationType: types.InvocationTypeRequestResponse,
		LogType:        types.LogTypeTail,
		Payload:        b,
	}
	if output, err := sam.Invoke(context.TODO(), input); err != nil {
		log.WithError(err).Error()
		return api.Err(err)
	} else {
		return api.OK(string(output.Payload))
	}
}

func campaigns(id string) (events.APIGatewayV2HTTPResponse, error) {
	b, _ := json.Marshal(map[string]interface{}{"campaigns": id})
	input := &faas.InvokeInput{
		FunctionName:   ptr.String("plumbus_accountHandler"),
		InvocationType: types.InvocationTypeRequestResponse,
		LogType:        types.LogTypeTail,
		Payload:        b,
	}
	if output, err := sam.Invoke(context.TODO(), input); err != nil {
		log.WithError(err).Error()
		return api.Err(err)
	} else {
		return api.OK(string(output.Payload))
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
