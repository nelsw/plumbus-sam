package rule

import (
	"github.com/google/uuid"
	"plumbus/pkg/model/campaign"
	"time"
)

const (
	handler = "plumbus_ruleHandler"
)

var table = "plumbus_rule"

func TableName() *string {
	return &table
}

func Handler() string {
	return handler
}

type Entity struct {

	// ID is the unique identifier and partition key.
	ID string `json:"id"`

	// Named is superfluous moniker for users.
	Named string `json:"name"`

	// Active is a flag used to determine rule inclusion.
	Active bool `json:"status"`

	// Conditions are statements which much prove true to
	Conditions []Condition `json:"conditions"`

	// Effect is the outcome of satisfactory rules on Ads.
	Effect campaign.Status `json:"effect"`

	// Nodes are a graph of Campaign ID's mapped by an Account ID.
	Nodes map[string][]string `json:"scope"`

	// Updated is the time this entity was last updated.
	Updated time.Time `json:"updated"`

	// Created is the time this entity was last created.
	Created time.Time `json:"created"`
}

func (e *Entity) PrePut() {
	now := time.Now().UTC()
	if e.Updated = now; e.ID == "" {
		e.ID = uuid.NewString()
		e.Created = now
	}
}

type LHS string

const (
	ROI    LHS = "ROI"
	Spend      = "SPEND"
	Profit     = "PROFIT"
)

type Op string

const (
	GT Op = ">"
	LT    = "<"
)

type Condition struct {
	LHS LHS     `json:"lhs"`
	Op  Op      `json:"op"`
	RHS float64 `json:"rhs"`
}

func (c Condition) Met(val float64) bool {
	return (c.Op == GT && val > c.RHS) || (c.Op == LT && val < c.RHS)
}
