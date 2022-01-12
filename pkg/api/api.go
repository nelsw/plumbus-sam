package api

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var headers = map[string]string{"Access-Control-Allow-Origin": "*"} // Required when CORS enabled in API Gateway.

func Empty() (events.APIGatewayV2HTTPResponse, error) {
	return worker(http.StatusNotFound, "")
}

func Err(err error) (events.APIGatewayV2HTTPResponse, error) {
	log.WithError(err).Error()
	return worker(http.StatusBadRequest, err.Error())
}

func Nada() (events.APIGatewayV2HTTPResponse, error) {
	return Err(errors.New("nothing handled"))
}

func OK(body string) (events.APIGatewayV2HTTPResponse, error) {
	return worker(http.StatusOK, body)
}

func K() (events.APIGatewayV2HTTPResponse, error) {
	return worker(http.StatusOK, "")
}

func OnlyOK(body string) (events.APIGatewayV2HTTPResponse, error) {
	return abbreviatedWorker(http.StatusOK, body)
}

func worker(code int, body string) (events.APIGatewayV2HTTPResponse, error) {
	log.WithFields(log.Fields{"code": code, "body": body}).Info()
	return events.APIGatewayV2HTTPResponse{Headers: headers, StatusCode: code, Body: body}, nil
}

func abbreviatedWorker(code int, body string) (events.APIGatewayV2HTTPResponse, error) {
	log.WithFields(log.Fields{"code": code, "body (len)": len(body)}).Info()
	return events.APIGatewayV2HTTPResponse{Headers: headers, StatusCode: code, Body: body}, nil
}

func NewRequestBytes(method string, params map[string]string) (out []byte) {
	req := NewRequest(method, params)
	out, _ = json.Marshal(&req)
	return
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