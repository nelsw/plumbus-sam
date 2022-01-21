package account

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"plumbus/pkg/model/campaign"
	"plumbus/pkg/util/compare"
	"strconv"
	"time"
)

var table = "plumbus_account"

func Table() string {
	return table
}

func TableName() *string {
	return &table
}

// ByName implements sort.Interface based on the Name field.
type ByName []Entity

func (n ByName) Len() int           { return len(n) }
func (n ByName) Swap(x, y int)      { n[x], n[y] = n[y], n[x] }
func (n ByName) Less(x, y int) bool { return compare.Strings(n[x].Named, n[y].Named) }

type Entity struct {

	// ID is the unique identifier assigned to this object by Facebook. It does not include the "act_" prefix.
	ID string

	// Named is the name of the account and effectively a magic string for aggregating data and producing KPI's. GL.
	// Name is a reserved keyword in DynamoDB.
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

func (e *Entity) Item() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"ID":       &types.AttributeValueMemberS{Value: e.ID},
		"Named":    &types.AttributeValueMemberS{Value: e.Named},
		"Stated":   &types.AttributeValueMemberN{Value: strconv.Itoa(e.Stated)},
		"Created":  &types.AttributeValueMemberS{Value: e.Created},
		"Included": &types.AttributeValueMemberBOOL{Value: e.Included},
	}
}

func (e *Entity) WriteRequest() types.WriteRequest {
	return types.WriteRequest{PutRequest: &types.PutRequest{Item: e.Item()}}
}

func (e *Entity) PutItemInput() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{Item: e.Item(), TableName: TableName()}
}
