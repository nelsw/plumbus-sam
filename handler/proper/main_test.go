package main

import (
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"testing"
)

func TestHandle(t *testing.T) {
	response, err := handle(events.APIGatewayV2HTTPRequest{})
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fail()
	}
}
