package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/model/fb"
	"plumbus/pkg/repo"
	"plumbus/pkg/util/logs"
)

var (
	campaignTable = "plumbus_fb_campaign"
)

func init() {
	logs.Init()
}

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": req}).Info()

	switch req.RequestContext.HTTP.Method {

	case http.MethodOptions:
		return api.K()

	case http.MethodPut:
		accountID := req.QueryStringParameters["account_id"]
		campaignID := req.QueryStringParameters["campaign_id"]
		status := req.QueryStringParameters["status"]
		return put(ctx, accountID, campaignID, fb.CampaignStatus(status))
	}

	return api.K()
}

func put(ctx context.Context, accountID, campaignID string, status fb.CampaignStatus) (events.APIGatewayV2HTTPResponse, error) {

	if campaignID != "" {
		if err := fb.UpdateCampaignStatus(campaignID, status); err != nil {
			return api.Err(err)
		} else if err = updateCampaignStatus(ctx, accountID, campaignID, status); err != nil {
			return api.Err(err)
		} else {
			return api.K()
		}
	}

	if cc, err := getCampaigns(ctx, accountID); err != nil {
		return api.Err(err)
	} else {
		for _, c := range cc {
			if err := fb.UpdateCampaignStatus(c.ID, status); err != nil {
				log.WithError(err).Error()
			} else if err = updateCampaignStatus(ctx, c.AccountID, c.ID, status); err != nil {
				log.WithError(err).Error()
			}
		}
	}

	return api.K()
}

func updateCampaignStatus(ctx context.Context, accountID, campaignID string, status fb.CampaignStatus) (err error) {

	in := &dynamodb.UpdateItemInput{
		TableName: &campaignTable,
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
				Value: status.String(),
			},
		},
		UpdateExpression: ptr.String("set Circ = :v1"),
	}

	_, err = repo.Update(ctx, in)
	return

}

type campaign struct {
	AccountID string `json:"account_id"`
	ID        string `json:"id"`
}

func getCampaigns(ctx context.Context, accountID string) (cc []campaign, err error) {

	in := &dynamodb.QueryInput{
		TableName:              &campaignTable,
		KeyConditionExpression: ptr.String("AccountID = :v1"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberS{Value: accountID},
		},
	}

	var out *dynamodb.QueryOutput
	if out, err = repo.Query(ctx, in); err != nil {
		return
	}

	err = attributevalue.UnmarshalListOfMaps(out.Items, &cc)
	return
}

func main() {
	lambda.Start(handle)
}
