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
	"plumbus/pkg/api"
	"plumbus/pkg/repo"
	"sync"
)

var (
	accountTable  = "plumbus_fb_account"
	campaignTable = "plumbus_fb_campaign"
)

type account struct {
	ID        string     `json:"account_id"`
	Name      string     `json:"name"`
	Level     string     `json:"level"`
	Campaigns []campaign `json:"children,omitempty"`
}

type campaign struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": request}).Info()

	return get(ctx)
}

func get(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	var aa []account
	if err := repo.Scan(&dynamodb.ScanInput{TableName: &accountTable}, &aa); err != nil {
		return api.Err(err)
	}

	var out []account
	var wg sync.WaitGroup
	for _, a := range aa {
		wg.Add(1)
		go func(a account) {
			defer wg.Done()
			var err error
			if a.Campaigns, err = campaigns(ctx, a.ID); err != nil {
				log.WithError(err).Error()
			}
			a.Level = "account"
			out = append(out, a)
		}(a)
	}
	wg.Wait()

	if data, err := json.Marshal(&out); err != nil {
		return api.Err(err)
	} else {
		return api.OnlyOK(string(data))
	}
}

func campaigns(ctx context.Context, accountID string) (cc []campaign, err error) {

	in := &dynamodb.QueryInput{
		TableName:              &campaignTable,
		KeyConditionExpression: ptr.String("AccountID = :v1"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberS{Value: accountID},
		},
	}

	var out *dynamodb.QueryOutput
	if out, err = repo.Query(ctx, in); err != nil {
		log.WithError(err).Error()
		return
	}

	if err = attributevalue.UnmarshalListOfMaps(out.Items, &cc); err != nil {
		log.WithError(err).Error()
	}

	return
}

func main() {
	lambda.Start(handle)
}
