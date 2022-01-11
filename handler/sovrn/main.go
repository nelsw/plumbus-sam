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

func init() {
	logs.Init()
}

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"req": req}).Info()

	if err := sovrn.Handle(ctx, req); err != nil {
		log.WithError(err).Error()
	}

	data := api.NewRequestBytes(http.MethodPut, map[string]string{"node": "root"})
	if _, err := sam.NewEvent(ctx, "plumbus_aggHandler", data); err != nil {
		return api.Err(err)
	}

	return api.OK("")
}

func main() {
	lambda.Start(handle)
}
