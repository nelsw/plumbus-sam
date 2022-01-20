package campaign

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shopspring/decimal"
	"plumbus/pkg/util/nums"
	"plumbus/pkg/util/pretty"
	"regexp"
	"strconv"
	"strings"
)

var (
	table        = "plumbus_fb_campaign"
	handler      = "plumbus_campaignHandler"
	statusRegexp = regexp.MustCompile("ACTIVE|PAUSED|DELETED|ARCHIVED")
)

func Table() string {
	return table
}

func TableName() *string {
	return &table
}

func Handler() string {
	return handler
}

type Status string

const (
	Active   Status = "ACTIVE"
	Paused   Status = "PAUSED"
	Deleted  Status = "DELETED"
	Archived Status = "ARCHIVED"
)

func (s Status) Param() string {
	return "status=" + s.String()
}

func (s Status) String() string {
	return string(s)
}

func (s Status) Validate() error {
	if str := s.String(); !statusRegexp.MatchString(str) {
		return errors.New(fmt.Sprintf("Invalid Status: [%s], must be ACTIVE, PAUSED, DELETED, or ARCHIVED", str))
	}
	return nil
}

type Entity struct {

	// AccountID is the partition key; It represents the Account which owns this campaign.
	AccountID string `json:"account_id"`

	// ID is the index key and unique identifier for this entity.
	ID string `json:"id"`

	// CampaignID is the index key and unique identifier for this entity.
	CampaignID string `json:"campaign_id,omitempty"`

	// Name is exactly what you think it is.
	Name string `json:"name"`

	// Circ is an awful workaround for an alleged keyword representing account status.
	Circ string `json:"status"`

	// Created is the time which this campaign was created.
	Created string `json:"created_time"`

	// Updated is the time which this campaign was updated.
	// Updating the campaign spend_cap, daily budget, or lifetime budget will not automatically update this field.
	Updated string `json:"updated_time"`

	// DailyBudget is the daily budget of the campaign
	DailyBudget string `json:"daily_budget,omitempty"`

	// BudgetRemaining is the
	BudgetRemaining string `json:"budget_remaining"`

	// Clicks are the number of clicks for campaigns ads.
	Clicks string `json:"clicks"`

	// Impressions represent the number of times campaign ads were on screen.
	Impressions string `json:"impressions"`

	// Spend is the estimated total amount of money spent on this campaign during its schedule.
	Spend string `json:"spend"`

	// CPC is the average cost for each click (all).
	CPC string `json:"cpc"`

	// CPP is the average cost to reach 1,000 people. This metric is estimated.
	CPP string `json:"cpp"`

	// CPM is the average cost for 1,000 impressions.
	CPM string `json:"cpm"`

	// CTR is the percentage of times people saw your ad and performed a click (all).
	CTR string `json:"ctr"`

	// UTM (Urchin Tracking Module) is a URL parameter used to track the effectiveness of online marketing campaigns.
	// In the case of this entity, the UTM is "sometimes" embedded in the campaign name.
	// In the scope of this system, the UTM is used as a UUID by SOVRN to track Ad revenue.
	UTM string `json:"utm"`

	// Revenue is the amount of (cumulative) revenue generated by campaign ads.
	Revenue float64 `json:"revenue"`

	// Profit is Revenue - Spend ... yeah dingus.
	Profit float64 `json:"profit"`

	// ROI is the return on investment for this campaign expressed as a decimal.
	// ROI decimal value is calculated by dividing Profit by Spend.
	// ROI percentage is calculated by multiplying the decimal value by 100. dummy.
	ROI float64 `json:"roi"`

	Formatted Formatted `json:"formatted"`
}

func (e *Entity) SetFormat() {
	dailyBudget, _ := decimal.NewFromString(e.DailyBudget)
	budgetRemaining, _ := decimal.NewFromString(e.BudgetRemaining)
	clicks, _ := decimal.NewFromString(e.Clicks)
	impressions, _ := decimal.NewFromString(e.Impressions)
	spend, _ := decimal.NewFromString(e.Spend)
	cpc, _ := decimal.NewFromString(e.CPC)
	cpp, _ := decimal.NewFromString(e.CPP)
	cpm, _ := decimal.NewFromString(e.CPM)
	ctr, _ := decimal.NewFromString(e.CTR)
	e.Formatted = Formatted{
		DailyBudget:     pretty.USD(dailyBudget, true),
		BudgetRemaining: pretty.USD(budgetRemaining),
		Clicks:          pretty.Int(clicks),
		Impressions:     pretty.Int(impressions),
		Spend:           pretty.USD(spend),
		CPC:             pretty.USD(cpc),
		CPP:             pretty.USD(cpp),
		CPM:             pretty.USD(cpm),
		CTR:             pretty.Percent(ctr, 2),
		Revenue:         pretty.USD(decimal.NewFromFloat(e.Revenue)),
		Profit:          pretty.USD(decimal.NewFromFloat(e.Profit)),
		ROI:             pretty.Percent(decimal.NewFromFloat(e.ROI), 0),
	}
}

func (e *Entity) Item() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"ID":              &types.AttributeValueMemberS{Value: e.ID},
		"AccountID":       &types.AttributeValueMemberS{Value: e.AccountID},
		"Name":            &types.AttributeValueMemberS{Value: e.Name},
		"Named":           &types.AttributeValueMemberS{Value: e.Name},
		"Circ":            &types.AttributeValueMemberS{Value: e.Circ},
		"Created":         &types.AttributeValueMemberS{Value: e.Created},
		"Updated":         &types.AttributeValueMemberS{Value: e.Updated},
		"DailyBudget":     &types.AttributeValueMemberS{Value: e.DailyBudget},
		"BudgetRemaining": &types.AttributeValueMemberS{Value: e.BudgetRemaining},
		"Clicks":          &types.AttributeValueMemberS{Value: e.Clicks},
		"Impressions":     &types.AttributeValueMemberS{Value: e.Impressions},
		"Spend":           &types.AttributeValueMemberS{Value: e.Spend},
		"CPC":             &types.AttributeValueMemberS{Value: e.CPC},
		"CPP":             &types.AttributeValueMemberS{Value: e.CPP},
		"CPM":             &types.AttributeValueMemberS{Value: e.CPM},
		"CTR":             &types.AttributeValueMemberS{Value: e.CTR},
		"UTM":             &types.AttributeValueMemberS{Value: e.UTM},
		"Revenue":         &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", e.Revenue)},
		"Profit":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", e.Profit)},
		"ROI":             &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", e.ROI)},
	}
}

func (e *Entity) WriteRequest() types.WriteRequest {
	return types.WriteRequest{PutRequest: &types.PutRequest{Item: e.Item()}}
}

func (e *Entity) PutItemInput() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{Item: e.Item(), TableName: TableName()}
}

func (e *Entity) Spent() float64 {
	f, _ := strconv.ParseFloat(e.Spend, 64)
	return f
}

func (e *Entity) SetUTM() {

	if chunks := strings.Split(e.Name, "("); len(chunks) > 1 {

		if chunks = strings.Split(chunks[1], ")"); len(chunks) > 1 {
			e.UTM = chunks[0]
			return
		}
	}

	if spaced := strings.Split(e.Name, " "); len(spaced) > 1 && nums.IsNumber(spaced[0]) {
		e.UTM = spaced[0]
		return
	}

	if scored := strings.Split(e.Name, "_"); len(scored) > 1 && nums.IsNumber(scored[0]) {
		e.UTM = scored[0]
		return
	}

	e.UTM = e.ID
}

type Formatted struct {
	DailyBudget     string `json:"daily_budget"`
	BudgetRemaining string `json:"budget_remaining"`
	Clicks          string `json:"clicks"`
	Impressions     string `json:"impressions"`
	Spend           string `json:"spend"`
	CPC             string `json:"cpc"`
	CPP             string `json:"cpp"`
	CPM             string `json:"cpm"`
	CTR             string `json:"ctr"`
	Revenue         string `json:"revenue"`
	Profit          string `json:"profit"`
	ROI             string `json:"roi"`
}
