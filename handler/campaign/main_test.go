package main

import (
	"net/http"
	"plumbus/pkg/sam"
	"plumbus/test"
	"testing"
)

func TestHandlePut(t *testing.T) {
	req := sam.NewRequest(http.MethodPut, map[string]string{"accountID": "2069406016568961"})
	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestHandlePatch(t *testing.T) {
	req := sam.NewRequest(http.MethodPatch, map[string]string{
		"accountID": "264100649065412",
		"ID":        "23850120462950705",
		"status":    "ACTIVE",
	})
	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestHandleGetByAccountID(t *testing.T) {
	par := map[string]string{"accountID": "1450566098533975"}
	req := sam.NewRequest(http.MethodGet, par)
	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestHandleGetByAccountIDAndCampaignIDS(t *testing.T) {
	par := map[string]string{
		"accountID":   "264100649065412",
		"campaignIDS": "23850120461840705,23850120462950705",
	}
	req := sam.NewRequest(http.MethodGet, par)
	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}
