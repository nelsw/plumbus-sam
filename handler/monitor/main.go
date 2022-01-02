package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/api"
	"plumbus/pkg/util/logs"
)

var sam *faas.Client

func init() {
	logs.Init()
	if cfg, err := config.LoadDefaultConfig(context.TODO()); err != nil {
		log.WithError(err).Fatal()
	} else {
		sam = faas.NewFromConfig(cfg)
	}
}

func handle(event events.CloudWatchEvent) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"event": event}).Info()

	// get all rules

	// for each rule

	return api.OK("")
}

func main() {
	lambda.Start(handle)
}
