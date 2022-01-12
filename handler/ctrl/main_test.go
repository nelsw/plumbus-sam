package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"plumbus/pkg/model/fb"
	"testing"
)

func TestHandleUpdateCampaignStatus(t *testing.T) {

	//p := map[string]string{
	//	"account_id":"264100649065412",
	//	"campaign_id": "23850061947590705",
	//	"status": fb.PausedCampaign.String(),
	//}
	//req := newRequest(http.MethodPut, p)
	//if res, _ := handle(context.TODO(), req); res.StatusCode != http.StatusOK {
	//	t.Error(res.StatusCode, res.Body)
	//}

	//p = map[string]string{
	//	"account_id":"264100649065412",
	//	"campaign_id": "23850061947590705",
	//	"status": fb.ActiveCampaign.String(),
	//}
	//req = newRequest(http.MethodPut, p)
	//if res, _ := handle(context.TODO(), req); res.StatusCode != http.StatusOK {
	//	t.Error(res.StatusCode, res.Body)
	//}
}

// 920821998512059

func TestHandleUpdateCampaignsStatus(t *testing.T) {

	p := map[string]string{
		"account_id": "920821998512059",
		"status":     fb.PausedCampaign.String(),
	}
	req := newRequest(http.MethodPut, p)
	if res, _ := handle(context.TODO(), req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}

	//p = map[string]string{
	//	"account_id":"920821998512059",
	//	"status": fb.ActiveCampaign.String(),
	//}
	//req = newRequest(http.MethodPut, p)
	//if res, _ := handle(context.TODO(), req); res.StatusCode != http.StatusOK {
	//	t.Error(res.StatusCode, res.Body)
	//}
}

func newRequest(method string, params map[string]string) events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		QueryStringParameters: params,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: method,
			},
		},
	}
}
