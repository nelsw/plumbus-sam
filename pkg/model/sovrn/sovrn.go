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

type value struct {
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

func (v value) writeRequest() (out types.WriteRequest, err error) {
	var item map[string]types.AttributeValue
	if item, err = attributevalue.MarshalMap(&v); err == nil {
		out.PutRequest = &types.PutRequest{Item: item}
	}
	return
}

func init() {
	logs.Init()
}

func Process(ctx context.Context, request events.APIGatewayV2HTTPRequest) (err error) {

	var pay payload
	if err = json.Unmarshal([]byte(strings.TrimLeft(request.Body, "attachment")), &pay); err != nil {
		log.WithError(err).Error("unable to interpret request.Body attachment payload")
		return
	}

	log.Trace("unmarshalled sovrn payload from request.Body attachment")

	var vv []value
	if err = json.Unmarshal([]byte(pay.Attachment.Data), &vv); err != nil {
		log.WithError(err).Error("unable to unmarshal payload attachment data into sovrn value slice")
		return err
	}

	log.WithFields(log.Fields{"size": len(vv)}).Trace("unmarshalled sovrn values from payload attachment data")

	// as each sovrn value represents a campaign,
	// we group sovrn values by (campaign) UTM
	// to sum and average value data points.
	groups := map[string][]value{}
	for _, v := range vv {
		if v.UTM == "" {
			log.Trace("value missing campaign id", v)
		} else if _, ok := groups[v.UTM]; ok {
			groups[v.UTM] = append(groups[v.UTM], v)
		} else {
			groups[v.UTM] = []value{v}
		}
	}

	log.WithFields(log.Fields{"size": len(vv)}).Trace("sovrn value groups")

	// here we perform said
	// summation and averaging,
	// before persisting to db.
	var r types.WriteRequest
	var rr []types.WriteRequest
	for utm, group := range groups {

		v := value{UTM: utm}

		for _, g := range group {
			v.Revenue += g.Revenue
			v.Impressions += g.Impressions
			v.Sessions += g.Sessions
			v.CTR += g.CTR
			v.PageViews += g.PageViews
		}

		v.CTR /= float64(len(group))

		if r, err = v.writeRequest(); err != nil {
			log.WithError(err).Error("sovrn value to write request")
		} else {
			rr = append(rr, r)
		}
	}

	if err = repo.BatchWriteItems(ctx, table, rr); err != nil {
		log.WithError(err).Error("sovrn value batch write items")
	} else {
		log.WithFields(log.Fields{"size": len(rr)}).Trace("sovrn value batch write items")
	}

	return
}
