package main

import (
	"encoding/json"
	"net/http"
	"plumbus/pkg/model/campaign"
	"plumbus/pkg/sam"
	"plumbus/pkg/util/pretty"
	"plumbus/test"
	"testing"
)

//func TestHandlePut(t *testing.T) {
//	req := sam.NewRequest(http.MethodPut, map[string]string{"accountID": "2989088818006204"})
//	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
//		t.Error(res.StatusCode, res.Body)
//	} else {
//		pretty.PrintJson(res.Body)
//	}
//}

func TestHandleGetByAccountID(t *testing.T) {
	par := map[string]string{"accountID": "1450566098533975"}
	req := sam.NewRequest(http.MethodGet, par)
	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	} else {
		var cc []campaign.Entity
		json.Unmarshal([]byte(res.Body), &cc)
		pretty.Print(cc)
	}
}

func TestHandleGetByAccountIDAndCampaignIDS(t *testing.T) {
	par := map[string]string{
		"accountID":   "544877570187911",
		"campaignIDS": "23849761526340551,23849761526450551",
	}
	req := sam.NewRequest(http.MethodGet, par)
	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}
