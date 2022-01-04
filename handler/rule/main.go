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
	Scope      string      `json:"scope"` // global for now
	Conditions []Condition `json:"conditions"`
	Action     string      `json:"action"` // enable / disable
	Status     bool        `json:"status"` // on / off
	Created    time.Time   `json:"created"`
	Updated    time.Time   `json:"updated"`
}

type Condition struct {
	Key      string  `json:"key"`      // roi % / spend
	Operator string  `json:"operator"` // gt, lt
	Value    float64 `json:"value"`
}

func init() {
	logs.Init()
}

func handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": request}).Info()

	method := request.RequestContext.HTTP.Method

	if method == http.MethodGet {

		var out interface{}
		if err := repo.ScanInputAndUnmarshal(&dynamodb.ScanInput{TableName: &table}, &out); err != nil {
			return api.Err(err)
		}

		bytes, _ := json.Marshal(out)
		return api.OK(string(bytes))
	}

	if method == http.MethodPut {

		var err error

		var rule Rule
		if err = json.Unmarshal([]byte(request.Body), &rule); err != nil {
			return api.Err(err)
		}

		if rule.ID == "" {
			rule.ID = uuid.NewString()
			rule.Created = time.Now().UTC()
		}

		rule.Updated = time.Now().UTC()

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

		id := request.QueryStringParameters["ID"]
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
