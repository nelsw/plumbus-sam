// Package provides functionality for updating and return campaign entity data on Facebook and the system database.
package main

import (
	"context"
	"encoding/json"
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
		accountID := req.QueryStringParameters["accountID"]
		if campaignIDS, ok := req.QueryStringParameters["campaignIDS"]; ok {
			return get(ctx, accountID, campaignIDS)
		} else {
			return query(ctx, accountID)
		}

	case http.MethodPatch:
		return patch(ctx, req)

	default:
		return api.Nada()
	}

}

func patch(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	status := campaign.Status(req.QueryStringParameters["status"])
	accountID := req.QueryStringParameters["accountID"]
	ID := req.QueryStringParameters["ID"]

	if ID != "" {
		if err := update(ctx, accountID, ID, status); err != nil {
			return api.Err(err)
		}
		return api.K()
	}

	in := &dynamodb.QueryInput{
		TableName:              campaign.TableName(),
		KeyConditionExpression: ptr.String("AccountID = :v1"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberS{
				Value: accountID,
			},
		},
	}

	var cc []campaign.Entity
	if out, err := repo.Query(ctx, in); err != nil {
		return api.Err(err)
	} else if err = attributevalue.UnmarshalListOfMaps(out.Items, &cc); err != nil {
		return api.Err(err)
	}

	for _, c := range cc {
		if err := update(ctx, accountID, c.ID, status); err != nil {
			return api.Err(err)
		}
	}

	return api.K()
}

func update(ctx context.Context, accountID, ID string, status campaign.Status) (err error) {

	param := map[string]interface{}{
		"node":   "campaign",
		"ID":     ID,
		"status": status,
	}

	data, _ := json.Marshal(param)
	if _, err = sam.NewReqRes(ctx, fb.Handler(), data); err != nil {
		log.WithError(err).Error()
		return
	}

	in := &dynamodb.UpdateItemInput{
		TableName: campaign.TableName(),
		Key: map[string]types.AttributeValue{
			"AccountID": &types.AttributeValueMemberS{
				Value: accountID,
			},
			"ID": &types.AttributeValueMemberS{
				Value: ID,
			},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberS{
				Value: status.String(),
			},
		},
		UpdateExpression: ptr.String("set Circ = :v1"),
	}

	if _, err = repo.Update(ctx, in); err != nil {
		log.WithError(err).Error()
	}

	return
}

// get queries the db for all campaigns where the given parameters equal the key expression attribute values.
func get(ctx context.Context, accountID string, campaignIDS string) (events.APIGatewayV2HTTPResponse, error) {

	var keys []map[string]types.AttributeValue

	for _, id := range strings.Split(campaignIDS, ",") {
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
