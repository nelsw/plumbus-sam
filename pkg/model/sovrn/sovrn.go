package sovrn

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var Table = "plumbus_fb_sovrn"

type Entity struct {
	UTM         string  `json:"UTM"`
	Revenue     float64 `json:"Revenue"`
	Impressions int     `json:"Impressions"`
	Sessions    int     `json:"Sessions"`
	CTR         float64 `json:"CTR"`
	PageViews   int     `json:"PageViews"`
}

type Payload struct {
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

func (v *Value) WriteRequest() (out types.WriteRequest, err error) {
	var item map[string]types.AttributeValue
	if item, err = attributevalue.MarshalMap(&v); err == nil {
		out.PutRequest = &types.PutRequest{Item: item}
	}
	return
}
