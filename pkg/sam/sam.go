package sam

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/util/logs"
)

var sam *faas.Client

func init() {
	logs.Init()
	if cfg, err := config.LoadDefaultConfig(context.Background()); err != nil {
		log.WithError(err).Fatal()
	} else {
		sam = faas.NewFromConfig(cfg)
	}
}

func NewRequest(method string, params map[string]string) events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		QueryStringParameters: params,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: method,
			},
		},
	}
}

func NewRequestBytes(method string, params map[string]string) (out []byte) {
	req := NewRequest(method, params)
	out, _ = json.Marshal(&req)
	return
}

func NewEvent(ctx context.Context, name string, data []byte) (out *faas.InvokeOutput, err error) {
	return invoke(ctx, &faas.InvokeInput{
		FunctionName:   &name,
		InvocationType: types.InvocationTypeEvent,
		Payload:        data,
	})
}

func NewReqRes(ctx context.Context, name string, data []byte) (out *faas.InvokeOutput, err error) {
	return invoke(ctx, &faas.InvokeInput{
		FunctionName:   &name,
		InvocationType: types.InvocationTypeRequestResponse,
		LogType:        types.LogTypeTail,
		Payload:        data,
	})
}

func invoke(ctx context.Context, in *faas.InvokeInput) (out *faas.InvokeOutput, err error) {
	if out, err = sam.Invoke(ctx, in); err != nil {
		log.WithError(err).Error()
	}
	return
}
