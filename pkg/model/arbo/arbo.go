package arbo

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strconv"
)

const imgHost = "https://dwyeew221rxbg.cloudfront.net/facebook_fu/"

var (
	table = "plumbus_arbo"
)

func Table() string {
	return table
}

func TableName() *string {
	return &table
}

type Payload struct {
	Data []Entity `json:"data"`
}

type Entity struct {
	Id           string      `json:"id"`
	Abid         string      `json:"abid"`
	Cid          string      `json:"cid"`
	PageId       string      `json:"page_id"`
	Nid          string      `json:"nid"`
	Checkbox     string      `json:"checkbox"`
	Status       string      `json:"status"`
	Network      []string    `json:"network"`
	TargetUrl    string      `json:"target_url"`
	Img          string      `json:"img"`
	Name         string      `json:"name"`
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
	Formatted    formatted   `json:"formatted"`
}

type formatted struct {
	Bid          string `json:"bid"`
	Budget       string `json:"budget"`
	Spend        string `json:"spend"`
	Clicks       string `json:"clicks"`
	Ctr          string `json:"ctr"`
	Ecpc         string `json:"ecpc"`
	Simpressions string `json:"simpressions"`
	Revenue      string `json:"revenue"`
	Profit       string `json:"profit"`
	Cpm          string `json:"cpm"`
	Rimpressions string `json:"rimpressions"`
	Rps          string `json:"rps"`
	Hrps         string `json:"hrps"`
	Roi          string `json:"roi"`
}

func (e *Entity) Item() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"ID":           &types.AttributeValueMemberS{Value: e.Cid},
		"UTM":          &types.AttributeValueMemberS{Value: e.Abid},
		"Named":        &types.AttributeValueMemberS{Value: e.Name},
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
	} else if s, err := strconv.ParseFloat(e.Spend.(string), 64); err != nil {
		return nil
	} else if r := e.Revenue.(float64); s == 0 {
		return r
	} else if p := r - s; r == 0 {
		return p
	} else {
		return p / s
	}
}

func (e *Entity) SetFormat() {

}

func (e *Entity) WriteRequest() types.WriteRequest {
	return types.WriteRequest{PutRequest: &types.PutRequest{Item: e.Item()}}
}

func (e *Entity) PutItemInput() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{Item: e.Item(), TableName: TableName()}
}
