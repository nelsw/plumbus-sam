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
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/model/arbo"
	"plumbus/pkg/model/campaign"
	"plumbus/pkg/model/fb"
	"plumbus/pkg/model/sovrn"
	"plumbus/pkg/repo"
	"plumbus/pkg/sam"
	"plumbus/pkg/util/logs"
	"plumbus/pkg/util/nums"
	"strings"
	"sync"
	"time"
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

	param := map[string]interface{}{
		"node": "campaigns",
		"ID":   req.QueryStringParameters["accountID"],
	}

	data, _ := json.Marshal(param)

	var err error
	var out *faas.InvokeOutput
	if out, err = sam.NewReqRes(ctx, fb.Handler, data); err != nil {
		log.WithError(err).Error()
		return api.Err(err)
	}

	var cc []campaign.Entity
	if err = json.Unmarshal(out.Payload, &cc); err != nil {
		log.WithError(err).Error()
		return api.Err(err)
	}

	var wg sync.WaitGroup
	var rr []types.WriteRequest
	for _, c := range cc {
		wg.Add(1)
		go func(c campaign.Entity) {
			defer wg.Done()
			c.SetUTM()
			c.SetFormat()
			if r, err := refresh(ctx, &c); err != nil {
				log.WithError(err).Warn()
			} else {
				rr = append(rr, r)
			}
		}(c)
	}

	wg.Wait()

	if err = repo.BatchWrite(ctx, campaign.Table, rr); err != nil {
		return api.Err(err)
	}

	return api.JSON(cc)
}

func refresh(ctx context.Context, c *campaign.Entity) (r types.WriteRequest, err error) {

	var a arbo.Entity
	var s sovrn.Entity

	if err = repo.Get(ctx, arbo.Table, "ID", c.ID, &a); err != nil {
		log.WithError(err).Warn()
	} else if err = repo.Get(ctx, sovrn.Table, "UTM", c.UTM, &s); err != nil {
		log.WithError(err).Warn()
	}

	if err != nil {
		log.Warn("errors getting arbo and sovrn data for " + c.ID)
		return
	}

	if a.ID != c.ID && s.UTM != c.UTM {
		log.Warn("arbo and sovrn entities do not match, panic!")
		return
	}

	if a.ID == c.ID {
		c.Revenue = nums.Float64(a.Revenue)
		c.Profit = nums.Float64(a.Profit)
		c.ROI = nums.Float64(a.Roi)
	} else {
		c.Revenue = s.Revenue
		c.Profit = c.Revenue - c.Spent()
		if c.Profit == 0 {
			c.ROI = 0
		} else if c.Spent() == 0 {
			c.ROI = 100
		} else if c.Revenue == 0 {
			c.ROI = -100
		} else {
			c.ROI = c.Profit / c.Spent() * 100
		}
	}

	c.Refreshed = time.Now()
	r = c.WriteRequest()

	return
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
	if _, err = sam.NewReqRes(ctx, fb.Handler, data); err != nil {
		log.WithError(err).Error()
		return
	}

	in := &dynamodb.UpdateItemInput{
		TableName: ptr.String(campaign.Table),
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
	}
	return query(ctx, accountID)
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
			campaign.Table: {
				Keys: keys,
			},
		},
	}

	var out *dynamodb.BatchGetItemOutput
	if out, err = repo.BatchGet(ctx, in); err != nil {
		log.WithError(err).Error()
	} else if err = attributevalue.UnmarshalListOfMaps(out.Responses[campaign.Table], &cc); err != nil {
		log.WithError(err).Error()
	} else {
		log.Trace("batch get for AccountID ", accountID, " found ", len(cc), " out of the given ", len(ids))
	}

	return
}

func query(ctx context.Context, accountID string) (cc []campaign.Entity, err error) {

	in := &dynamodb.QueryInput{
		TableName:              ptr.String(campaign.Table),
		KeyConditionExpression: ptr.String("AccountID = :v1"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberS{
				Value: accountID,
			},
		},
	}

	var out *dynamodb.QueryOutput
	if out, err = repo.Query(ctx, in); err != nil {
		log.WithError(err).Error()
	} else if err = attributevalue.UnmarshalListOfMaps(out.Items, &cc); err != nil {
		log.WithError(err).Error()
	} else {
		log.Trace("query for AccountID ", accountID, " found ", len(cc))
	}

	return
}

func main() {
	lambda.Start(handle)
}
