package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/api"
	"plumbus/pkg/repo"
	"plumbus/pkg/util"
	"plumbus/pkg/util/logs"
	"strings"
)

var (
	table = "plumbus_fb_revenue"
)

func init() {
	logs.Init()
}

type Impressions struct {
	Campaign string  `json:"impressions.utm_campaign"`
	AdSet    string  `json:"impressions.utm_adset"`
	Revenue  float64 `json:"impressions.estimated_revenue"`
	Account  string  `json:"sovrn_account"`
}

func (i Impressions) toPutItemInput() *dynamodb.PutItemInput {

	var rev = "0.0"
	if i.Revenue > 0 {
		rev = util.FloatToDecimal(i.Revenue)
	}
	rev = strings.ReplaceAll(rev, ",", "")

	return &dynamodb.PutItemInput{
		TableName: &table,
		Item: map[string]types.AttributeValue{
			"campaign": &types.AttributeValueMemberS{Value: i.Campaign},
			"adset":    &types.AttributeValueMemberS{Value: i.AdSet},
			"revenue":  &types.AttributeValueMemberN{Value: rev},
			"account":  &types.AttributeValueMemberS{Value: i.Account},
		},
	}
}

func handle(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"req": request}).Info()

	var err error

	var payload struct {
		Attachment struct {
			Data string `json:"data"`
		} `json:"attachment"`
	}
	if err = json.Unmarshal([]byte(strings.TrimLeft(request.Body, "attachment")), &payload); err != nil {
		log.WithError(err).Error()
		return api.OK("")
	}

	var impressions []Impressions
	if err = json.Unmarshal([]byte(payload.Attachment.Data), &impressions); err != nil {
		log.WithError(err).Error()
		return api.OK("")
	}

	account := impressions[0].Account

	var out []interface{}
	if err = repo.ScanInputAndUnmarshal(&dynamodb.ScanInput{TableName: &table}, &out); err != nil {
		log.WithError(err).Error()
		return api.OK("")
	}

	for _, o := range out {
		if m, ok := o.(map[string]interface{}); !ok {
			fmt.Println("want type map[string]interface{};  got %T", o)
		} else {
			acct := fmt.Sprintf("%v", m["account"])
			if acct == "" || acct == account {
				key := fmt.Sprintf("%v", m["campaign"])
				if err = repo.DelByEntry(table, "campaign", key); err != nil {
					log.WithError(err).Error("while deleting", key)
				}
			}
		}
	}

	for _, impression := range impressions {
		if err = repo.Put(impression.toPutItemInput()); err != nil {
			log.WithError(err).Error()
		}
	}

	return api.OK("")
}

func main() {
	lambda.Start(handle)
}
