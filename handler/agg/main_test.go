package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"plumbus/pkg/util"
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
	if res, _ := handle(ctx, newRequest(http.MethodPut, accountParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestPutCampaigns(t *testing.T) {
	if res, _ := handle(ctx, newRequest(http.MethodPut, campaignParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestPostRoot(t *testing.T) {
	if res, _ := handle(ctx, newRequest(http.MethodPost, rootParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestPostAccounts(t *testing.T) {
	if res, _ := handle(ctx, newRequest(http.MethodPost, accountParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestPostCampaigns(t *testing.T) {
	if res, _ := handle(ctx, newRequest(http.MethodPost, campaignParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestGetRoot(t *testing.T) {
	if res, _ := handle(ctx, newRequest(http.MethodGet, rootParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestGetAccounts(t *testing.T) {
	if res, _ := handle(ctx, newRequest(http.MethodGet, accountParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	} else {
		util.PrettyPrint(res.Body)
	}
}

func TestGetCampaigns(t *testing.T) {
	if res, _ := handle(ctx, newRequest(http.MethodGet, campaignIDParam)); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
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
