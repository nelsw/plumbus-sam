package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/repo"
	"plumbus/pkg/util/logs"
	"time"
)

var (
	table     = "plumbus_fb_rule"
	scanInput = dynamodb.ScanInput{TableName: &table}
)

type Rule struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Status      bool        `json:"status"` // on / off
	Conditions  []Condition `json:"conditions"`
	Action      bool        `json:"action"` // enable / disable
	Created     time.Time   `json:"created"`
	Updated     time.Time   `json:"updated"`
	AccountIDS  []string    `json:"account_ids"`
	CampaignIDS []string    `json:"campaign_ids"`

	Scope []string `json:"scope"` // deprecated
}

type Target string

const (
	targetROI   = "ROI"
	targetSpend = "SPEND"
)

type Operator string

const (
	operatorGT = ">"
	operatorLT = "<"
)

type Condition struct {
	ID       string    `json:"id"`
	Target   Target    `json:"target"`   // roi % / $ spend
	Operator Operator  `json:"operator"` // gt, lt
	Value    float64   `json:"value"`    // roi % / $ spend
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

func init() {
	logs.Init()
}

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": req}).Info()

	switch req.RequestContext.HTTP.Method {

	case http.MethodOptions:
		return api.K()

	case http.MethodGet:
		return handleGet(ctx)

	case http.MethodPut:
		return handlePut(ctx, []byte(req.Body))

	case http.MethodDelete:
		return handleDelete(ctx, req.QueryStringParameters["id"])

	case http.MethodPost:
		return handlePost(ctx)

	default:
		return api.Err(errors.New("nothing handled"))
	}
}

func handleGet(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	var out []Rule
	if err := repo.Scan(ctx, &scanInput, &out); err != nil {
		return api.Err(err)
	}

	var data []byte
	data, _ = json.Marshal(out)
	return api.OK(string(data))
}

func handlePut(ctx context.Context, data []byte) (events.APIGatewayV2HTTPResponse, error) {

	var rule Rule
	if err := json.Unmarshal(data, &rule); err != nil {
		return api.Err(err)
	}

	now := time.Now().UTC()
	if rule.Updated = now; rule.ID == "" {
		rule.ID = uuid.NewString()
		rule.Created = now
	}

	for index, condition := range rule.Conditions {
		if rule.Conditions[index].Updated = now; condition.ID == "" {
			rule.Conditions[index].ID = uuid.NewString()
			rule.Conditions[index].Created = now
		}
	}

	if item, err := attributevalue.MarshalMap(&rule); err != nil {
		return api.Err(err)
	} else if err = repo.Put(ctx, &dynamodb.PutItemInput{Item: item, TableName: &table}); err != nil {
		return api.Err(err)
	} else {
		bytes, _ := json.Marshal(&rule)
		return api.OK(string(bytes))
	}
}

func handleDelete(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {

	in := &dynamodb.DeleteItemInput{
		TableName: &table,
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{
				Value: id,
			},
		},
	}

	if err := repo.DeleteItem(ctx, in); err != nil {
		return api.Err(err)
	}

	return api.OK("")
}

func handlePost(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	fmt.Println("not yet running") // todo - assess
	return api.OK("")
}

func main() {
	lambda.Start(handle)
}
