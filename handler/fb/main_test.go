package main

import (
	"plumbus/pkg/model/campaign"
	"plumbus/pkg/util/pretty"
	"plumbus/test"
	"testing"
)

func TestHandleAccounts(t *testing.T) {
	if _, err := handle(test.CTX, map[string]interface{}{"node": "accounts"}); err != nil {
		t.Error(err)
	}
}

func TestHandleCampaigns(t *testing.T) {
	param := map[string]interface{}{
		"node": "campaigns",
		"ID":   "264100649065412",
	}
	if res, err := handle(test.CTX, param); err != nil {
		t.Error(err)
	} else {
		pretty.Print(res)
	}
}

func TestHandleCampaignStatusUpdate(t *testing.T) {

	param := map[string]interface{}{
		"node":   "campaign",
		"ID":     "23850120461840705",
		"status": campaign.Paused,
	}

	if _, err := handle(test.CTX, param); err != nil {
		t.Error(err)
	}
}
