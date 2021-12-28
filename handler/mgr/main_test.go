package main

import (
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"testing"
)

func TestHandle(t *testing.T) {

	res, _ := handle(events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			RouteKey:     "",
			AccountID:    "",
			Stage:        "",
			RequestID:    "",
			Authorizer:   nil,
			APIID:        "",
			DomainName:   "",
			DomainPrefix: "",
			Time:         "",
			TimeEpoch:    0,
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method:    http.MethodGet,
				Path:      "",
				Protocol:  "",
				SourceIP:  "",
				UserAgent: "",
			},
			Authentication: events.APIGatewayV2HTTPRequestContextAuthentication{},
		},
	})
	if res.StatusCode != http.StatusOK {
		t.Fail()
	}

}
