package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"testing"
)

func TestHandle(t *testing.T) {

	ctx := context.TODO()
	evt := events.CloudWatchEvent{}

	res, _ := handle(ctx, evt)
	if res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
		return
	}

}
