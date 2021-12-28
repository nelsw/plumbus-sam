package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
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

	var out *dynamodb.ScanOutput
	if out, err = db.Scan(context.TODO(), &dynamodb.ScanInput{TableName: &tableName}); err != nil {
		return api.Err(err)
	}

	var old []struct {
		Campaign string  `json:"campaign"`
		Revenue  float64 `json:"revenue"`
	}

	if err = attributevalue.UnmarshalListOfMaps(out.Items, &old); err != nil {
		return api.Err(err)
	}

	for _, o := range old {
		dii := &dynamodb.DeleteItemInput{
			TableName: &tableName,
			Key:       map[string]types.AttributeValue{"campaign": &types.AttributeValueMemberS{Value: o.Campaign}},
		}
		if _, err = db.DeleteItem(context.TODO(), dii); err != nil {
			fmt.Println(err)
		}
	}
	total := 0.0
	totals := map[string]float64{}
	for _, d := range data {
		if v, b := totals[d.Campaign]; b {
			totals[d.Campaign] = v + d.Revenue
		} else {
			totals[d.Campaign] = d.Revenue
		}
		total += d.Revenue
	}

	fmt.Println(total)

	for c, r := range totals {
		var rev = "0.0"
		if r > 0 {
			rev = util.FloatToDecimal(r)
		}
		rev = strings.ReplaceAll(rev, ",", "")
		put := &dynamodb.PutItemInput{
			TableName: &tableName,
			Item: map[string]types.AttributeValue{
				"campaign": &types.AttributeValueMemberS{Value: c},
				"revenue":  &types.AttributeValueMemberN{Value: rev},
			},
		}
		if _, err = db.PutItem(context.TODO(), put); err != nil {
			fmt.Println(c, r, err)
		}
	}

	return api.OK("")
}

func main() {
	lambda.Start(handle)
}
