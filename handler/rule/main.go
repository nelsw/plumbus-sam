package main

import (
	"context"
	"encoding/json"
	"errors"
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

var table = "plumbus_fb_rule"

type Rule struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Status     bool        `json:"status"` // on / off
	Scope      []string    `json:"scope"`  // account id's
	Conditions []Condition `json:"conditions"`
	Action     bool        `json:"action"` // enable / disable
	Created    time.Time   `json:"created"`
	Updated    time.Time   `json:"updated"`
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

func handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": request}).Info()

	method := request.RequestContext.HTTP.Method

	if method == http.MethodOptions {
		return api.OK("")
	}

	if method == http.MethodGet {

		var out interface{}
		if err := repo.ScanInputAndUnmarshal(&dynamodb.ScanInput{TableName: &table}, &out); err != nil {
			return api.Err(err)
		}

		var rules []Rule

		bytes, _ := json.Marshal(out)
		if err := json.Unmarshal(bytes, &rules); err != nil {
			return api.Err(err)
		}

		bytes, _ = json.Marshal(rules)
		return api.OK(string(bytes))
	}

	if method == http.MethodPut {

		var err error

		var rule Rule
		if err = json.Unmarshal([]byte(request.Body), &rule); err != nil {
			return api.Err(err)
		}

		now := time.Now().UTC()
		rule.Updated = now
		if rule.ID == "" {
			rule.ID = uuid.NewString()
			rule.Created = now
		}

		for index, condition := range rule.Conditions {
			rule.Conditions[index].Updated = now
			if condition.ID == "" {
				rule.Conditions[index].ID = uuid.NewString()
				rule.Conditions[index].Created = now
			}
		}

		var item map[string]types.AttributeValue
		if item, err = attributevalue.MarshalMap(&rule); err != nil {
			return api.Err(err)
		}

		if err = repo.Put(&dynamodb.PutItemInput{Item: item, TableName: &table}); err != nil {
			return api.Err(err)
		}

		bytes, _ := json.Marshal(&rule)
		return api.OK(string(bytes))
	}

	if method == http.MethodDelete {

		id := request.QueryStringParameters["id"]
		if err := repo.DelByEntry(table, "ID", id); err != nil {
			return api.Err(err)
		}

		return api.OK("")
	}

	return api.Err(errors.New("nothing handled"))
}

func main() {
	lambda.Start(handle)
}
