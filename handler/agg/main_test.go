package main

import (
	"context"
	"net/http"
	"plumbus/pkg/sam"
	"testing"
)

var (
	ctx             = context.TODO()
	rootParam       = map[string]string{"node": "root"}
	accountParam    = map[string]string{"node": "account"}
	campaignParam   = map[string]string{"node": "campaign"}
	campaignIDParam = map[string]string{"node": "campaign", "id": "414566673354941"}
)

func TestPutAccounts(t *testing.T) {
	if res, _ := handle(ctx, sam.NewRequest(http.MethodPut, accountParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestPutCampaigns(t *testing.T) {
	if res, _ := handle(ctx, sam.NewRequest(http.MethodPut, campaignParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestPostRoot(t *testing.T) {
	if res, _ := handle(ctx, sam.NewRequest(http.MethodPost, rootParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestPostAccounts(t *testing.T) {
	if res, _ := handle(ctx, sam.NewRequest(http.MethodPost, accountParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestPostCampaigns(t *testing.T) {
	if res, _ := handle(ctx, sam.NewRequest(http.MethodPost, campaignParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestGetRoot(t *testing.T) {
	if res, _ := handle(ctx, sam.NewRequest(http.MethodGet, rootParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestGetAccounts(t *testing.T) {
	if res, _ := handle(ctx, sam.NewRequest(http.MethodGet, accountParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestGetCampaigns(t *testing.T) {
	if res, _ := handle(ctx, sam.NewRequest(http.MethodGet, campaignIDParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}
