package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"os"
	"plumbus/pkg/api"
	"plumbus/pkg/util"
	"strings"
)

var (
	db        *dynamodb.Client
	tableName = "plumbus_fb_revenue"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	if cfg, err := config.LoadDefaultConfig(context.TODO()); err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		panic(err)
	} else {
		db = dynamodb.NewFromConfig(cfg)
	}
}

type Payload struct {
	Attachment struct {
		Data string `json:"data"`
	} `json:"attachment"`
}

type Impressions struct {
	Campaign string  `json:"impressions.utm_campaign"`
	Revenue  float64 `json:"impressions.estimated_revenue"`
}

func (i Impressions) toPutItemInput() *dynamodb.PutItemInput {
	var rev = "0.0"
	if i.Revenue > 0 {
		rev = util.FloatToDecimal(i.Revenue)
	}
	rev = strings.ReplaceAll(rev, ",", "")
	return &dynamodb.PutItemInput{
		TableName: &tableName,
		Item: map[string]types.AttributeValue{
			"campaign": &types.AttributeValueMemberS{Value: i.Campaign},
			"revenue":  &types.AttributeValueMemberN{Value: rev},
		},
	}
}

func handle(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"req": request}).Info()

	var err error

	var payload Payload
	if err = json.Unmarshal([]byte(strings.TrimLeft(request.Body, "attachment")), &payload); err != nil {
		panic(err)
	}

	var data []Impressions
	if err = json.Unmarshal([]byte(payload.Attachment.Data), &data); err != nil {
		panic(err)
	}

	for _, d := range data {
		if _, err = db.PutItem(context.TODO(), d.toPutItemInput()); err != nil {
			fmt.Println(d)
			fmt.Println(err)
		}
	}

	return api.OK("")
}

func main() {
	lambda.Start(handle)
}
