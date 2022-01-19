// Package provides functionality for handling accounts to "ignore".
// Ignored accounts are those that exist and excluded by processing.
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
	log "github.com/sirupsen/logrus"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/repo"
	"plumbus/pkg/util/logs"
)

var table = "plumbus_ignored_account"

type account struct {
	ID string `json:"account_id"` // does not include "act_" prefix
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
		return get(ctx, req.QueryStringParameters["form"])

	case http.MethodDelete:
		return del(ctx, req.QueryStringParameters["id"])

	case http.MethodPut:
		return put(ctx, req.QueryStringParameters["id"])

	default:
		return api.Nada()
	}
}

func get(ctx context.Context, form string) (events.APIGatewayV2HTTPResponse, error) {

	var err error

	var aa []account
	if err = repo.Scan(ctx, &dynamodb.ScanInput{TableName: &table}, &aa); err != nil {
		return api.Err(err)
	}

	if form == "arr" || form == "slice" {
		data, _ := json.Marshal(&aa)
		return api.Data(data)
	}

	if form == "map" {
		aaa := map[string]account{}
		for _, a := range aa {
			aaa[a.ID] = a
		}
		data, _ := json.Marshal(&aaa)
		return api.Data(data)
	}

	return api.Err(errors.New("unrecognized form: " + form))
}

func del(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {

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

	return api.K()
}

func put(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {
	if item, err := attributevalue.MarshalMap(&account{id}); err != nil {
		return api.Err(err)
	} else if err = repo.Put(ctx, &dynamodb.PutItemInput{Item: item, TableName: &table}); err != nil {
		return api.Err(err)
	} else {
		return api.K()
	}
}

func main() {
	lambda.Start(handle)
}
