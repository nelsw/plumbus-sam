package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"plumbus/pkg/util"
	"testing"
)

func TestHandle(t *testing.T) {
	if res, _ := handle(context.TODO(), events.APIGatewayV2HTTPRequest{}); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	} else {
		var aa []account
		if err := json.Unmarshal([]byte(res.Body), &aa); err != nil {
			t.Error(err)
		} else {
			util.PrettyPrint(aa)
		}
	}
}
