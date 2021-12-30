package svc

import (
	"context"
	"encoding/json"
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/util/logs"
)

var (
	input = &faas.InvokeInput{
		FunctionName:   ptr.String("plumbus_accountHandler"),
		InvocationType: types.InvocationTypeRequestResponse,
		LogType:        types.LogTypeTail,
	}
)

func init() {
	logs.Init()
}

func Accounts() (out []interface{}, err error) {
	var output *faas.InvokeOutput
	if output, err = sam.Invoke(context.WithValue(context.TODO(), "accounts", ""), input); err != nil {
		log.WithError(err).Error()
	} else if err = json.Unmarshal(output.Payload, &out); err != nil {
		log.WithError(err).Error()
	}
	return
}

func AdAccountsToIgnoreMap() (out map[string]interface{}, err error) {

	var invokeOutput *faas.InvokeOutput
	if invokeOutput, err = sam.Invoke(context.TODO(), input); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var payload []map[string]string
	if err = json.Unmarshal(invokeOutput.Payload, &payload); err != nil {
		log.WithError(err).Error()
	} else {
		for _, account := range payload {
			out[account["account_id"]] = nil
		}
	}

	return
}

func AdAccountsToIgnoreSlice() (out []string, err error) {

	var invokeOutput *faas.InvokeOutput
	if invokeOutput, err = sam.Invoke(context.TODO(), input); err != nil {
		log.WithError(err).Error()
		return
	}

	var payload []map[string]string
	if err = json.Unmarshal(invokeOutput.Payload, &payload); err != nil {
		log.WithError(err).Error()
	} else {
		for _, account := range payload {
			out = append(out, account["account_id"])
		}
	}

	return
}
