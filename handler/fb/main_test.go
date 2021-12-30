package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"
	"net/http"
	"plumbus/pkg/svc"
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
	var data []AdAccount
	_ = json.Unmarshal([]byte(res.Body), &data)

	var err error
	accountsToIgnoreMap := map[string]interface{}{}
	if accountsToIgnoreMap, err = svc.AdAccountsToIgnoreMap(); err != nil {
		log.WithError(err).Error()
	}

	for _, d := range data {
		if _, ok := accountsToIgnoreMap[d.AccountID]; ok {
			fmt.Println("fack")
		}
	}
}

//func TestHandleCamps(t *testing.T) {
//	res, _ := handle(events.APIGatewayV2HTTPRequest{
//		QueryStringParameters: map[string]string{
//			"domain": "camps",
//			"id":     "564715394630862",
//		},
//	})
//	if res.StatusCode != http.StatusOK {
//		t.Fail()
//	}
//}
//
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
