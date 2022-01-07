package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"plumbus/pkg/util"
	"testing"
	"time"
)

func TestHandle(t *testing.T) {

	alpha := time.Now()
	res, _ := handle(context.TODO(), events.APIGatewayV2HTTPRequest{})
	omega := time.Now()
	fmt.Println("duration ", omega.Sub(alpha))

	if res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
		return
	}

	var accounts []Account
	if err := json.Unmarshal([]byte(res.Body), &accounts); err != nil {
		t.Error(res.StatusCode, res.Body)
		return
	}

	util.PrettyPrint(accounts)
}
