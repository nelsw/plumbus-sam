package campaign

import (
	log "github.com/sirupsen/logrus"
	"strconv"
)

var (
	table   = "plumbus_fb_campaign"
	handler = "plumbus_campaignHandler"
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

func (s Status) Status() string {
	return "&status=" + string(s)
}

func (s Status) String() string {
	return string(s)
}

type Entity struct {
	AccountID string `json:"account_id"`
	ID        string `json:"id"`

	Named           string `json:"name"`
	Circ            string `json:"status"`
	DailyBudget     string `json:"daily_budget"`
	RemainingBudget string `json:"remaining_budget"`
	Created         string `json:"created_time"`
	Updated         string `json:"updated_time"`

	Clicks      string `json:"clicks"`
	Impressions string `json:"impressions"`
	Spend       string `json:"spend"`
	CPC         string `json:"cpc"` // The average cost for each click (all).
	CPP         string `json:"cpp"` // The average cost to reach 1,000 people. This metric is estimated.
	CPM         string `json:"cpm"` // The average cost for 1,000 impressions.
	CTR         string `json:"ctr"` // The percentage of times people saw your ad and performed a click (all).

	UTM     string  `json:"utm"`
	Revenue float64 `json:"revenue"`
	Profit  float64 `json:"profit"`
	ROI     float64 `json:"roi"`
}

func (e *Entity) Spent() float64 {
	f, err := strconv.ParseFloat(e.Spend, 64)
	if err != nil {
		log.WithError(err).Warn("unable to parse float: ", e.Spend)
	}
	return f // 0 or parsed value yeah dummy
}
