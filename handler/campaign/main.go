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
	"plumbus/pkg/model/sovrn"
	"plumbus/pkg/repo"
	"plumbus/pkg/sam"
	"plumbus/pkg/util/logs"
	"strconv"
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
		return get(ctx, req)
	case http.MethodPatch:
		return patch(ctx, req)
	case http.MethodPut:
		return put(ctx, req)
	default:
		return api.Nada()
	}
}

func put(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	ID := req.QueryStringParameters["accountID"]
	data, _ := json.Marshal(map[string]interface{}{"node": "campaigns", "ID": ID})
	if out, err := sam.NewReqRes(ctx, fb.Handler(), data); err != nil {
		log.WithError(err).Error()
		return api.Err(err)
	} else {

		var cc []campaign.Entity
		if err = json.Unmarshal(out.Payload, &cc); err != nil {
			log.WithError(err).Error()
			return api.Err(err)
		}

		var rr []types.WriteRequest
		for _, c := range cc {

			var x sovrn.Entity
			if err = repo.Get(ctx, sovrn.Table(), "UTM", c.UTM, &x); err != nil {
				log.WithError(err).Error()
				return api.Err(err)
			}

			if x == (sovrn.Entity{}) {
				continue
			}

			rev := x.Revenue
			spe := c.Spent()
			pro := x.Revenue - spe

			var roi float64
			if pro != 0 && spe == 0 {
				roi = pro
			} else if pro == 0 && spe != 0 {
				roi = spe * -1
			} else if pro != 0 && spe != 0 {
				roi = pro / spe
			}
			c.Revenue = rev
			c.Profit = pro
			c.ROI = roi

			if c.Impressions == "" {
				c.Impressions = strconv.Itoa(x.Impressions)
			}

			rr = append(rr, c.WriteRequest())
		}

		if err = repo.BatchWrite(ctx, campaign.Table(), rr); err != nil {
			return api.Err(err)
		}

		return api.JSON(cc)
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

	if cc, err := query(ctx, accountID); err != nil {
		return api.Err(err)
	} else {
		for _, c := range cc {
			if err := update(ctx, accountID, c.ID, status); err != nil {
				return api.Err(err)
			}
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
func get(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if out, err := find(ctx, req); err != nil {
		return api.Err(err)
	} else if len(out) == 0 {
		return api.Empty()
	} else {
		for i := range out {
			out[i].SetFormat()
		}
		return api.JSON(out)
	}
}

func find(ctx context.Context, req events.APIGatewayV2HTTPRequest) ([]campaign.Entity, error) {
	accountID := req.QueryStringParameters["accountID"]
	if campaignIDS, ok := req.QueryStringParameters["campaignIDS"]; ok {
		return batch(ctx, accountID, strings.Split(campaignIDS, ","))
	} else {
		return query(ctx, accountID)
	}
}

func batch(ctx context.Context, accountID string, ids []string) (cc []campaign.Entity, err error) {

	var keys []map[string]types.AttributeValue
	for _, id := range ids {
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

	var out *dynamodb.BatchGetItemOutput
	if out, err = repo.BatchGet(ctx, in); err != nil {
		log.WithError(err).Error()
		return
	} else if err = attributevalue.UnmarshalListOfMaps(out.Responses[campaign.Table()], &cc); err != nil {
		log.WithError(err).Error()
		return
	}

	log.Trace("batch get for AccountID ", accountID, " found ", len(cc), " out of the given ", len(ids))
	return
}

func query(ctx context.Context, accountID string) (cc []campaign.Entity, err error) {

	in := &dynamodb.QueryInput{
		TableName:              campaign.TableName(),
		KeyConditionExpression: ptr.String("AccountID = :v1"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberS{Value: accountID},
		},
	}

	var out *dynamodb.QueryOutput
	if out, err = repo.Query(ctx, in); err != nil {
		log.WithError(err).Error()
		return
	} else if err = attributevalue.UnmarshalListOfMaps(out.Items, &cc); err != nil {
		log.WithError(err).Error()
		return
	}

	log.Trace("query for AccountID ", accountID, " found ", len(cc))
	return
}

func main() {
	lambda.Start(handle)
}
