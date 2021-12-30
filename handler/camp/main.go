package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/fb"
	"plumbus/pkg/util/logs"
)

func init() {
	logs.Init()
}

func handle(ctx context.Context) (out map[string]interface{}, err error) {

	/*
		return info for a single campaign_id
	*/
	if value := ctx.Value("campaign_id"); value != nil {
		if out, err = getCampaignByID(value); err != nil {
			log.WithError(err).Error()
		}
	}

	if value := ctx.Value("account_id"); value != nil {
		if out, err = getCampaignsByAccountID(value); err != nil {
			log.WithError(err).Error()
		}
	}

	return
}

func getCampaignByID(value interface{}) (got map[string]interface{}, err error) {
	id := fmt.Sprintf("%v", value)
	if got, err = fb.Campaign(id).Marketing().GET(); err != nil {
		log.WithError(err).Error()
	}
	return
}

func getCampaignsByAccountID(value interface{}) (got map[string]interface{}, err error) {
	id := fmt.Sprintf("%v", value)
	if got, err = fb.Campaigns(id).Marketing().GET(); err != nil {
		log.WithError(err).Error()
	}
	return
}

func main() {
	lambda.Start(handle)
}
