package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go/ptr"
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

func handle(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"req": request}).Info()

	domain := request.QueryStringParameters["domain"]
	id := request.QueryStringParameters["id"]

	switch domain {
	case "acts":
		return accounts(ctx)
	case "camps":
		return campaigns(ctx, id)
	default:
		return api.Err(errors.New("unrecognized domain: " + domain))
	}
}

func accounts(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	b, _ := json.Marshal(map[string]interface{}{"accounts": true})
	in := &faas.InvokeInput{
		FunctionName:   ptr.String("plumbus_accountHandler"),
		InvocationType: types.InvocationTypeRequestResponse,
		LogType:        types.LogTypeTail,
		Payload:        b,
	}
	if out, err := sam.Invoke(ctx, in); err != nil {
		return api.Err(err)
	} else {
		return api.OK(string(out.Payload))
	}
}

func campaigns(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {
	b, _ := json.Marshal(map[string]interface{}{"campaigns": id})
	in := &faas.InvokeInput{
		FunctionName:   ptr.String("plumbus_accountHandler"),
		InvocationType: types.InvocationTypeRequestResponse,
		LogType:        types.LogTypeTail,
		Payload:        b,
	}
	if out, err := sam.Invoke(ctx, in); err != nil {
		return api.Err(err)
	} else {
		return api.OK(string(out.Payload))
	}
}

func main() {
	lambda.Start(handle)
}
