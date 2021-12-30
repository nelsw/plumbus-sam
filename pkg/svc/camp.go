package svc

import (
	"context"
	"encoding/json"
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	log "github.com/sirupsen/logrus"
)

func Campaigns(accountID string) (out []interface{}, err error) {
	var output *faas.InvokeOutput
	if output, err = sam.Invoke(context.WithValue(context.TODO(), "account_id", accountID), input); err != nil {
		log.WithError(err).Error()
	} else if err = json.Unmarshal(output.Payload, &out); err != nil {
		log.WithError(err).Error()
	}
	return
}
