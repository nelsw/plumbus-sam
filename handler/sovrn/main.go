package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/model/sovrn"
	"plumbus/pkg/sam"
	"plumbus/pkg/util/logs"
)

var data = api.NewRequestBytes(http.MethodPost, map[string]string{"node": "root"})

func init() {
	logs.Init()
}

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": req}).Info()

	if err := sovrn.Process(ctx, req); err != nil {
		log.WithError(err).Error("while processing sovrn request")
	} else if _, err = sam.NewEvent(ctx, "plumbus_aggHandler", data); err != nil {
		log.WithError(err).Error("while invoking agg event")
	}

	// as sovrn is actively hitting this webhook,
	// we always return a 200 from this handler
	// to communicate successful delivery.
	return api.K()
}

func main() {
	lambda.Start(handle)
}
