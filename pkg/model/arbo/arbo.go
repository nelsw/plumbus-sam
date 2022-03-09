package arbo

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"plumbus/pkg/util/nums"
)

const (
	imgHost = "https://dwyeew221rxbg.cloudfront.net/facebook_fu/"
	Table   = "plumbus_arbo"
	Handler = "plumbus_arboHandler"
)

type Payload struct {
	Data []Entity `json:"data"`
}

type Entity struct {
	Id           string      `json:"id"`
	ID           string      `json:"cid"`
	UTM          string      `json:"abid"`
	PageId       string      `json:"page_id"`
	Nid          string      `json:"nid"`
	Checkbox     string      `json:"checkbox"`
	Stated       string      `json:"status"`
	Network      []string    `json:"network"`
	TargetUrl    string      `json:"target_url"`
	Img          string      `json:"img"`
	Named        string      `json:"name"`
	Bid          string      `json:"bid"`
	Budget       string      `json:"budget"`
	Buyer        string      `json:"buyer"`
	Spend        interface{} `json:"spend"`
	Clicks       interface{} `json:"clicks"`
	Ctr          interface{} `json:"ctr"`
	Ecpc         interface{} `json:"ecpc"`
	Simpressions interface{} `json:"simpressions"`
	Revenue      interface{} `json:"revenue"`
	Profit       interface{} `json:"profit"`
	Cpm          interface{} `json:"cpm"`
	Rimpressions interface{} `json:"rimpressions"`
	Rps          interface{} `json:"rps"`
	Hrps         interface{} `json:"hrps"`
	Roi          interface{} `json:"roi"`
	Stime        string      `json:"stime"`
}

func (e *Entity) item() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"ID":           &types.AttributeValueMemberS{Value: e.ID},
		"UTM":          &types.AttributeValueMemberS{Value: e.UTM},
		"Named":        &types.AttributeValueMemberS{Value: e.Named},
		"Img":          &types.AttributeValueMemberS{Value: fmt.Sprintf("%s%s", imgHost, e.Img)},
		"Bid":          attributeValue(e.Bid),
		"Budget":       attributeValue(e.Budget),
		"Spend":        attributeValue(e.Spend),
		"Clicks":       attributeValue(e.Clicks),
		"CTR":          attributeValue(e.Ctr),
		"eCPC":         attributeValue(e.Ecpc),
		"sImpressions": attributeValue(e.Simpressions),
		"Revenue":      attributeValue(e.Revenue),
		"Profit":       attributeValue(e.Profit),
		"CPM":          attributeValue(e.Cpm),
		"rImpressions": attributeValue(e.Rimpressions),
		"RPS":          attributeValue(e.Rps),
		"hRPS":         attributeValue(e.Hrps),
		"ROI":          attributeValue(e.roi()),
		"sTime":        &types.AttributeValueMemberS{Value: e.Stime},
	}
}

func attributeValue(v interface{}) types.AttributeValue {
	if v == nil {
		return &types.AttributeValueMemberNULL{Value: true}
	} else {
		return &types.AttributeValueMemberS{Value: fmt.Sprintf("%v", v)}
	}
}

func (e *Entity) roi() interface{} {
	if e.Roi != nil && e.Roi != "" {
		return e.Roi
	} else if e.Spend == nil || e.Revenue == nil {
		return nil
	} else {
		spend := nums.Float64(e.Spend)
		revenue := nums.Float64(e.Revenue)
		profit := revenue - spend
		if profit == 0 || (spend == 0 && revenue == 0) {
			return 0
		} else if spend == 0 {
			return 100
		} else if revenue == 0 {
			return -100
		} else {
			return profit / spend * 100
		}
	}
}

func (e *Entity) WriteRequest() types.WriteRequest {
	return types.WriteRequest{PutRequest: &types.PutRequest{Item: e.item()}}
}
