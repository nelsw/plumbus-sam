package fb

import (
	"plumbus/pkg/util"
	"strings"
)

type Payload struct {
	Data []interface{} `json:"data"`
	Page struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type Account struct {
	ID        string     `json:"id"`
	AccountID string     `json:"account_id"`
	Name      string     `json:"name"`
	Status    int        `json:"account_status"`
	Created   string     `json:"created_time"`
	Campaigns []Campaign `json:"children,omitempty"`
}

type Campaign struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	DailyBudget     string `json:"daily_budget,omitempty"`
	RemainingBudget string `json:"remaining_budget,omitempty"`
	Created         string `json:"created_time"`
	Updated         string `json:"updated_time"`
}

type Insight struct {
}

func (c Campaign) UTM() string {
	utm := getParenWrappedSovrnID(c.ID)
	if utm == "" {
		if spaced := strings.Split(c.Name, " "); len(spaced) > 1 && util.IsNumber(spaced[0]) {
			utm = spaced[0]
		} else if scored := strings.Split(c.Name, "_"); len(scored) > 1 && util.IsNumber(scored[0]) {
			utm = scored[0]
		} else {
			utm = c.ID
		}
	}
	return utm
}

func getParenWrappedSovrnID(s string) string {
	if chunks := strings.Split(s, "("); len(chunks) > 1 {
		chunk := chunks[1]
		chunks = strings.Split(chunk, ")")
		if len(chunks) > 1 {
			return chunks[0]
		}
	}
	return ""
}
