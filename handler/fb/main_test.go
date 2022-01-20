package main

import (
	"plumbus/pkg/model/campaign"
	"plumbus/pkg/util/pretty"
	"plumbus/test"
	"testing"
)

func TestHandleAccounts(t *testing.T) {
	if _, err := handle(test.CTX, map[string]interface{}{"node": "account"}); err != nil {
		t.Error(err)
	}
}

func TestHandleCampaigns(t *testing.T) {
	param := map[string]interface{}{
		"node": "campaigns",
		"ID":   "833316934027692",
	}
	if res, err := handle(test.CTX, param); err != nil {
		t.Error(err)
	} else {
		pretty.PrintJson(res)
	}
}

func TestHandleCampaignStatusUpdate(t *testing.T) {

	param := map[string]interface{}{
		"node":   "campaign",
		"ID":     "23849344181080687",
		"status": campaign.Paused,
	}

	if _, err := handle(test.CTX, param); err != nil {
		t.Error(err)
	}
}
