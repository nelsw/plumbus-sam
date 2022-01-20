// Package provides functionality for updating and return account entity data.
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
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/model/campaign"
	"plumbus/pkg/model/fb"
	"plumbus/pkg/repo"
	"plumbus/pkg/sam"
	"plumbus/pkg/util/logs"
	"strings"
)

func init() {
	logs.Init()
}

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": req}).Info()

	switch req.RequestContext.HTTP.Method {

	case http.MethodOptions:
		return api.K()

	case http.MethodGet:
		if accountID, ok := req.QueryStringParameters["accountID"]; !ok {
			return api.Err(errors.New("missing accountID"))
		} else if campaignIDS, ok := req.QueryStringParameters["campaignIDS"]; !ok {
			return query(ctx, accountID)
		} else {
			return get(ctx, accountID, strings.Split(campaignIDS, ","))
		}

	case http.MethodPatch:
		if accountID, ok := req.QueryStringParameters["accountID"]; !ok {
			return api.Err(errors.New("missing accountID"))
		} else if campaignID, ok := req.QueryStringParameters["campaignID"]; !ok {
			return api.Err(errors.New("missing campaignID"))
		} else if status, ok := req.QueryStringParameters["status"]; !ok {
			return api.Err(errors.New("missing status"))
		} else {
			return patch(ctx, accountID, campaignID, status)
		}

	default:
		return api.Nada()
	}

}

func patch(ctx context.Context, accountID, campaignID, status string) (events.APIGatewayV2HTTPResponse, error) {
	data, _ := json.Marshal(map[string]interface{}{"node": "campaign", "id": campaignID, "status": status})
	if _, err := sam.NewReqRes(ctx, fb.Handler(), data); err != nil {
		return api.Err(err)
	}
	in := &dynamodb.UpdateItemInput{
		TableName: campaign.TableName(),
		Key: map[string]types.AttributeValue{
			"AccountID": &types.AttributeValueMemberS{
				Value: accountID,
			},
			"ID": &types.AttributeValueMemberS{
				Value: campaignID,
			},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberS{
				Value: status,
			},
		},
		UpdateExpression: ptr.String("set Circ = :v1"),
	}

	if _, err := repo.Update(ctx, in); err != nil {
		return api.Err(err)
	}

	return api.K()
}

// get queries the db for all campaigns where the given parameters equal the key expression attribute values.
func get(ctx context.Context, accountID string, campaignIDS []string) (events.APIGatewayV2HTTPResponse, error) {

	var keys []map[string]types.AttributeValue

	for _, id := range campaignIDS {
		key := map[string]types.AttributeValue{
			"AccountID": &types.AttributeValueMemberS{
				Value: accountID,
			},
			"ID": &types.AttributeValueMemberS{
				Value: id,
			},
		}
		keys = append(keys, key)
	}

	in := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			campaign.Table(): {
				Keys: keys,
			},
		},
	}

	var cc []campaign.Entity
	if out, err := repo.BatchGet(ctx, in); err != nil {
		return api.Err(err)
	} else if err = attributevalue.UnmarshalListOfMaps(out.Responses[campaign.Table()], &cc); err != nil {
		return api.Err(err)
	} else if len(cc) == 0 {
		return api.Empty()
	} else {
		return api.JSON(cc)
	}
}

func query(ctx context.Context, accountID string) (events.APIGatewayV2HTTPResponse, error) {

	in := &dynamodb.QueryInput{
		TableName:              campaign.TableName(),
		KeyConditionExpression: ptr.String("AccountID = :v1"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberS{Value: accountID},
		},
	}

	var cc []campaign.Entity
	if out, err := repo.Query(ctx, in); err != nil {
		return api.Err(err)
	} else if err = attributevalue.UnmarshalListOfMaps(out.Items, &cc); err != nil {
		return api.Err(err)
	} else if len(cc) == 0 {
		return api.Empty()
	} else {
		return api.JSON(cc)
	}
}

func main() {
	lambda.Start(handle)
}
