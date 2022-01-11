package sam

import (
	"context"
	"encoding/json"
	"fmt"
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

func AdAccountsToIgnoreMap() (out map[string]interface{}, err error) {

	var invokeOutput *faas.InvokeOutput
	if invokeOutput, err = sam.Invoke(context.TODO(), input); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	out = map[string]interface{}{}

	var payload map[string]interface{}
	if err = json.Unmarshal(invokeOutput.Payload, &payload); err != nil {
		log.WithError(err).Error()
	} else {
		for k, v := range payload {
			out[k] = v
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

	fmt.Println(string(invokeOutput.Payload))

	var payload map[string]interface{}
	if err = json.Unmarshal(invokeOutput.Payload, &payload); err != nil {
		log.WithError(err).Error()
	} else {
		for k := range payload {
			out = append(out, k)
		}
	}

	return
}
