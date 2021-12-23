package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"plumbus/pkg/util"
	"testing"
)

func TestHandleActs(t *testing.T) {
	res, _ := handle(events.APIGatewayV2HTTPRequest{QueryStringParameters: map[string]string{"domain": "acts"}})
	if res.StatusCode != http.StatusOK {
		t.Fail()
	}
	var data []Data
	_ = json.Unmarshal([]byte(res.Body), &data)
	util.PrettyPrint(data)
}

func TestHandleCamps(t *testing.T) {
	res, _ := handle(events.APIGatewayV2HTTPRequest{QueryStringParameters: map[string]string{"domain": "camps"}})
	if res.StatusCode != http.StatusOK {
		t.Fail()
	}
}

func TestHandleSets(t *testing.T) {
	res, _ := handle(events.APIGatewayV2HTTPRequest{QueryStringParameters: map[string]string{"domain": "sets"}})
	if res.StatusCode != http.StatusOK {
		t.Fail()
	}
}

func TestHandleAds(t *testing.T) {
	res, _ := handle(events.APIGatewayV2HTTPRequest{QueryStringParameters: map[string]string{"domain": "ads"}})
	if res.StatusCode != http.StatusOK {
		t.Fail()
	}
}
