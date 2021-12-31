package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/api"
	"plumbus/pkg/repo"
	"plumbus/pkg/util/logs"
	"strconv"
	"strings"
	"time"
)

var table = "plumbus_fb_sovrn"

func init() {
	logs.Init()
}

type Payload struct {
	Attachment struct {
		Data string `json:"data"`
	} `json:"attachment"`
}

type Value struct {
	Campaign    string  `json:"impressions.utm_campaign"`
	Revenue     float64 `json:"impressions.estimated_revenue"`
	Impressions int     `json:"impressions.total_ad_impressions"`
	Sessions    int     `json:"impressions.total_sessions"`
	CTR         float64 `json:"impressions.click_through_rate"`
	PageViews   int     `json:"impressions.total_page_views"`
}

func handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"req": request}).Info()

	var err error

	var payload Payload
	if err = json.Unmarshal([]byte(strings.TrimLeft(request.Body, "attachment")), &payload); err != nil {
		log.WithError(err).Error()
		return api.OK("")
	}

	var values []Value
	if err = json.Unmarshal([]byte(payload.Attachment.Data), &values); err != nil {
		log.WithError(err).Error()
		return api.OK("")
	}

	groups := map[string][]Value{}
	for _, value := range values {

		if value.Campaign == "" {
			log.Trace("value missing campaign id", value)
			value.Campaign = time.Now().UTC().Format(time.RFC3339)
		}

		if _, ok := groups[value.Campaign]; ok {
			groups[value.Campaign] = append(groups[value.Campaign], value)
		} else {
			groups[value.Campaign] = []Value{value}
		}
	}

	var requests []types.WriteRequest
	for id, group := range groups {

		var rev, ctr float64
		var imp, ses, pge int

		for _, g := range group {
			rev += g.Revenue
			ctr += g.CTR
			imp += g.Impressions
			ses += g.Sessions
			pge += g.PageViews
		}

		ctr /= float64(len(group))

		requests = append(requests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					"campaign":    &types.AttributeValueMemberS{Value: id},
					"revenue":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", rev)},
					"impressions": &types.AttributeValueMemberN{Value: strconv.Itoa(imp)},
					"sessions":    &types.AttributeValueMemberN{Value: strconv.Itoa(ses)},
					"ctr":         &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", ctr)},
					"page_views":  &types.AttributeValueMemberN{Value: strconv.Itoa(pge)},
				},
			},
		})
	}

	var in *dynamodb.BatchWriteItemInput
	for _, chunk := range chunkSlice(requests, 25) {
		in = &dynamodb.BatchWriteItemInput{RequestItems: map[string][]types.WriteRequest{table: chunk}}
		if _, err = repo.BatchWriteItem(ctx, in); err != nil {
			log.WithError(err).Error()
		}
	}

	return api.OK("")
}

func chunkSlice(slice []types.WriteRequest, size int) [][]types.WriteRequest {

	var chunks [][]types.WriteRequest
	var end int

	for i := 0; i < len(slice); i += size {

		if end = i + size; end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

func main() {
	lambda.Start(handle)
}
