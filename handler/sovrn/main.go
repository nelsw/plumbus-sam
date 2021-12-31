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
	"strconv"
	"strings"
)

var (
	table = "plumbus_fb_revenue"
)

func init() {
	logs.Init()
}

type Impressions struct {
	Campaign    string  `json:"impressions.utm_campaign"`
	AdSet       string  `json:"impressions.utm_adset"`
	SubID       string  `json:"impressions.subid"`
	Revenue     float64 `json:"impressions.estimated_revenue"`
	Impressions int     `json:"impressions.total_ad_impressions"`
	SessionsRPM float64 `json:"impressions.sessions_rpm"`
	CTR         float64 `json:"impressions.click_through_rate"`
	PageRPM     float64 `json:"impressions.page_rpm"`
	Account     string  `json:"sovrn_account"`
}

func (i *Impressions) toPutItemInput() *dynamodb.PutItemInput {
	//if item, err := attributevalue.MarshalMap(i); err != nil {
	//	return nil
	//} else {
	//	return &dynamodb.PutItemInput{TableName: &table, Item: item}
	//}

	return &dynamodb.PutItemInput{
		TableName: &table,
		Item: map[string]types.AttributeValue{
			"impressions.utm_campaign":         &types.AttributeValueMemberS{Value: i.Campaign},
			"impressions.utm_adset":            &types.AttributeValueMemberS{Value: i.AdSet},
			"impressions.subid":                &types.AttributeValueMemberS{Value: i.SubID},
			"impressions.estimated_revenue":    &types.AttributeValueMemberN{Value: util.FloatToDecimal(i.Revenue)},
			"impressions.total_ad_impressions": &types.AttributeValueMemberN{Value: strconv.Itoa(i.Impressions)},
			"impressions.sessions_rpm":         &types.AttributeValueMemberN{Value: util.FloatToDecimal(i.SessionsRPM)},
			"impressions.click_through_rate":   &types.AttributeValueMemberS{Value: util.FloatToDecimal(i.CTR)},
			"impressions.page_rpm":             &types.AttributeValueMemberS{Value: util.FloatToDecimal(i.PageRPM)},
			"sovrn_account":                    &types.AttributeValueMemberS{Value: i.Account},
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

	fmt.Println("impressions", impressions)
	util.PrettyPrint(impressions)

	var out interface{}
	if err = repo.ScanInputAndUnmarshal(&dynamodb.ScanInput{TableName: &table}, &out); err != nil {
		log.WithError(err).Error()
		return api.OK("")
	}

	fmt.Println(out)

	if len(out.([]interface{})) > 0 {
		for _, o := range out.([]interface{}) {
			if m, ok := o.(map[string]interface{}); !ok {
				fmt.Println("want type map[string]interface{};  got ", o)
			} else {
				acct := fmt.Sprintf("%v", m["account"])
				if acct == "" || acct == impressions[0].Account {
					key := fmt.Sprintf("%v", m["campaign"])
					if err = repo.DelByEntry(table, "campaign", key); err != nil {
						log.WithError(err).Error("while deleting", key)
					}
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
