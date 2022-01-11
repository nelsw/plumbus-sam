package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"testing"
)

var (
	ctx = context.TODO()
)

func TestPutAccounts(t *testing.T) {

	req := newRequest(http.MethodPut, map[string]string{"node": "account"})

	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
		return
	}
}

func TestGetAccounts(t *testing.T) {

	req := newRequest(http.MethodGet, map[string]string{"node": "account"})

	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
		return
	}
}

func TestPutCampaigns(t *testing.T) {

	req := newRequest(http.MethodPut, map[string]string{"node": "campaign"})

	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
		return
	}
}

func TestGetCampaigns(t *testing.T) {

	req := newRequest(http.MethodGet, map[string]string{"node": "campaign", "id": "414566673354941"})

	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
		return
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
