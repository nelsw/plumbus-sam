package svc

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	sam         *faas.Client
	ctx         = context.TODO()
	invokeInput = faas.InvokeInput{
		FunctionName: aws.String("plumbus_accountHandler"),
		LogType:      "Tail",
	}
)

func init() {

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
		ForceColors:   true,
	})

	if cfg, err := config.LoadDefaultConfig(ctx); err != nil {
		log.WithError(err).Fatal()
	} else {
		sam = faas.NewFromConfig(cfg)
	}
}

func AdAccountsToIgnoreMap() (map[string]interface{}, error) {

	var err error

	var invokeOutput *faas.InvokeOutput
	if invokeOutput, err = sam.Invoke(ctx, &invokeInput); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var payload []map[string]string
	if err = json.Unmarshal(invokeOutput.Payload, &payload); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	out := map[string]interface{}{}
	for _, account := range payload {
		out[account["account_id"]] = nil
	}

	return out, nil
}

func AdAccountsToIgnoreSlice() ([]string, error) {

	var err error

	var invokeOutput *faas.InvokeOutput
	if invokeOutput, err = sam.Invoke(ctx, &invokeInput); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var payload []map[string]string
	if err = json.Unmarshal(invokeOutput.Payload, &payload); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var out []string
	for _, account := range payload {
		out = append(out, account["account_id"])
	}

	return out, nil
}
