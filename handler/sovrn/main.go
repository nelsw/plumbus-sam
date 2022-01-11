package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/api"
	"plumbus/pkg/model/sovrn"
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

	// go get latest fb data

	return api.OK("")
}

func main() {
	lambda.Start(handle)
}
