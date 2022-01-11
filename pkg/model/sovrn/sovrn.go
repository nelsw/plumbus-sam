package sovrn

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/repo"
	"plumbus/pkg/util/logs"
	"strings"
)

const table = "plumbus_fb_sovrn"

type payload struct {
	Attachment struct {
		Data string `json:"data"`
	} `json:"attachment"`
}

type Value struct {
	UTM         string  `json:"impressions.utm_campaign"`
	Revenue     float64 `json:"impressions.estimated_revenue"`
	Impressions int     `json:"impressions.total_ad_impressions"`
	Sessions    int     `json:"impressions.total_sessions"`
	CTR         float64 `json:"impressions.click_through_rate"`
	PageViews   int     `json:"impressions.total_page_views"`
}

type Entity struct {
	UTM         string  `json:"UTM"`
	Revenue     float64 `json:"Revenue"`
	Impressions int     `json:"Impressions"`
	Sessions    int     `json:"Sessions"`
	CTR         float64 `json:"CTR"`
	PageViews   int     `json:"PageViews"`
}

func (e Entity) Table() string {
	return table
}

func (v Value) putRequest() (*types.PutRequest, error) {
	if item, err := attributevalue.MarshalMap(&v); err != nil {
		return nil, err
	} else {
		return &types.PutRequest{Item: item}, nil
	}
}

func init() {
	logs.Init()
}

func Handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (err error) {

	var p payload
	if err = json.Unmarshal([]byte(strings.TrimLeft(request.Body, "attachment")), &p); err != nil {
		log.WithError(err).Error()
		return
	}

	var vv []Value
	if err = json.Unmarshal([]byte(p.Attachment.Data), &vv); err != nil {
		log.WithError(err).Error()
		return
	}

	groups := map[string][]Value{}
	for _, value := range vv {
		if value.UTM == "" {
			log.Trace("value missing campaign id", value)
		} else if _, ok := groups[value.UTM]; ok {
			groups[value.UTM] = append(groups[value.UTM], value)
		} else {
			groups[value.UTM] = []Value{value}
		}
	}

	var r *types.PutRequest
	var rr []types.WriteRequest
	for utm, group := range groups {

		v := Value{UTM: utm}

		for _, g := range group {
			v.Revenue += g.Revenue
			v.Impressions += g.Impressions
			v.Sessions += g.Sessions
			v.CTR += g.CTR
			v.PageViews += g.PageViews
		}

		v.CTR /= float64(len(group))

		if r, err = v.putRequest(); err != nil {
			log.WithError(err).Error()
			return
		}

		rr = append(rr, types.WriteRequest{PutRequest: r})
	}

	for _, c := range chunkSlice(rr, 25) {
		if _, err = repo.BatchWriteItems(ctx, table, c); err != nil {
			log.WithError(err).Error()
			return
		}
	}

	return
}

func chunkSlice(slice []types.WriteRequest, size int) (chunks [][]types.WriteRequest) {
	var end int
	for i := 0; i < len(slice); i += size {
		if end = i + size; end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return
}
