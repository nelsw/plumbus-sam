package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"testing"
)

func TestHandleActs(t *testing.T) {

	res, _ := handle(events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{
			"domain": "acts",
		},
	})
	if res.StatusCode != http.StatusOK {
		t.Fail()
	}

	fmt.Println(res.Body)

}

func TestHandleCamps(t *testing.T) {
	res, _ := handle(events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{
			"domain": "camps",
			"id":     "564715394630862",
		},
	})
	if res.StatusCode != http.StatusOK {
		t.Fail()
	}
}

//func TestHandleSets(t *testing.T) {
//	res, _ := handle(events.APIGatewayV2HTTPRequest{QueryStringParameters: map[string]string{"domain": "sets"}})
//	if res.StatusCode != http.StatusOK {
//		t.Fail()
//	}
//}
//
//func TestHandleAds(t *testing.T) {
//	res, _ := handle(events.APIGatewayV2HTTPRequest{QueryStringParameters: map[string]string{"domain": "ads"}})
//	if res.StatusCode != http.StatusOK {
//		t.Fail()
//	}
//}
