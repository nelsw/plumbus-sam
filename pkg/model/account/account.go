package account

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/ptr"
	"plumbus/pkg/model/campaign"
	"plumbus/pkg/util/compare"
	"plumbus/pkg/util/pretty"
	"strconv"
	"time"
)

const (
	Table   = "plumbus_account"
	Handler = "plumbus_accountHandler"
)

// ByName implements sort.Interface based on the Name field.
type ByName []Entity

func (n ByName) Len() int           { return len(n) }
func (n ByName) Swap(x, y int)      { n[x], n[y] = n[y], n[x] }
func (n ByName) Less(x, y int) bool { return compare.Strings(n[x].Named, n[y].Named) }

type Entity struct {

	// ID is the unique identifier assigned to this object by Facebook. It does not include the "act_" prefix.
	ID string

	// Named is the name of the account as Name is a reserved keyword in DynamoDB.
	Named string

	// Created is an ISO 8601 formatted datetime representing the instant this account was created in Facebook.
	// We store the value "as-is", but also provide an RFC 3339 format when marshaling.
	Created string

	// Stated is the encoded status value of this account as it exists in Facebook.
	// Status is a reserved keyword in DynamoDB.
	Stated int

	// Included is a flag used by Plumbus to determine which accounts should be considered when executing rules.
	Included bool

	// Children are the campaign entities of any status which are owned by this account.
	Children []campaign.Entity

	Performance Performance
}

type Performance struct {
	Spend float64 `json:"spend"`

	SpendStr string `json:"spend_str"`

	Revenue float64 `json:"revenue"`

	RevenueStr string `json:"revenue_str"`

	Profit float64 `json:"profit"`

	ProfitStr string `json:"profit_str"`

	ROI float64 `json:"roi"`

	ROIStr string `json:"roi_str"`

	Active int `json:"active"`

	ActiveStr string `json:"active_str"`
}

func (p *Performance) SetFormat() {
	p.SpendStr = pretty.USD(p.Spend)
	p.RevenueStr = pretty.USD(p.Revenue)
	p.ProfitStr = pretty.USD(p.Profit)
	p.ROIStr = pretty.Percent(p.ROI, 0)
	p.ActiveStr = pretty.Int(p.Active)
}

func (e *Entity) MarshalJSON() (data []byte, err error) {

	var status string
	switch e.Stated {
	case 1:
		status = "Active"
	case 2:
		status = "Disabled"
	case 3:
		status = "Unsettled"
	case 7:
		status = "PendingRiskReview"
	case 8:
		status = "PendingSettlement"
	case 9:
		status = "InGracePeriod"
	case 100:
		status = "PendingClosure"
	case 101:
		status = "Closed"
	case 201:
		status = "AnyActive"
	case 202:
		status = "AnyClosed"
	default:
		status = "Unknown Status: " + strconv.Itoa(e.Stated)
	}

	var created time.Time
	created, err = time.Parse("2006-01-02T15:04:05-0700", e.Created)

	if e.Children != nil && len(e.Children) > 0 {
		for _, c := range e.Children {
			if c.Stated == campaign.Active {
				e.Performance.Active += 1
			}
			e.Performance.Spend += c.Spent()
			e.Performance.Revenue += c.Revenue
			e.Performance.Profit += c.Profit
			e.Performance.ROI += c.ROI
		}
		e.Performance.SetFormat()
	}

	return json.Marshal(map[string]interface{}{
		"id":             e.ID,
		"account_id":     e.ID,
		"name":           e.Named,
		"account_status": e.Stated,
		"created_time":   e.Created,
		"included":       e.Included,
		"status":         status,
		"created":        created,
		"children":       e.Children,
		"performance":    e.Performance,
	})
}

func (e *Entity) UnmarshalJSON(data []byte) (err error) {

	var m map[string]*json.RawMessage
	if err = json.Unmarshal(data, &m); err != nil {
		return
	}

	for k, v := range m {
		switch k {
		case "account_id":
			err = json.Unmarshal(*v, &e.ID)
		case "name":
			err = json.Unmarshal(*v, &e.Named)
		case "account_status":
			err = json.Unmarshal(*v, &e.Stated)
		case "created_time":
			err = json.Unmarshal(*v, &e.Created)
		case "children":
			if e.Children != nil {
				err = json.Unmarshal(*v, &e.Children)
			}
		}
		if err != nil {
			return
		}
	}

	return
}

func (e *Entity) item() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"ID":       &types.AttributeValueMemberS{Value: e.ID},
		"Named":    &types.AttributeValueMemberS{Value: e.Named},
		"Stated":   &types.AttributeValueMemberN{Value: strconv.Itoa(e.Stated)},
		"Created":  &types.AttributeValueMemberS{Value: e.Created},
		"Included": &types.AttributeValueMemberBOOL{Value: e.Included},
	}
}

func (e *Entity) WriteRequest() types.WriteRequest {
	return types.WriteRequest{PutRequest: &types.PutRequest{Item: e.item()}}
}

func (e *Entity) PutItemInput() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{Item: e.item(), TableName: ptr.String(Table)}
}
