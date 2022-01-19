package main

import (
	"context"
	"net/http"
	"plumbus/pkg/sam"
	"testing"
)

var ctx = context.TODO()

func TestHandleGetByAccountID(t *testing.T) {
	par := map[string]string{"accountID": "414566673354941"}
	req := sam.NewRequest(http.MethodGet, par)
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestHandleGetByAccountIDAndCampaignIDS(t *testing.T) {
	par := map[string]string{
		"accountID":   "544877570187911",
		"campaignIDS": "23849761526340551,23849761526450551",
	}
	req := sam.NewRequest(http.MethodGet, par)
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}
