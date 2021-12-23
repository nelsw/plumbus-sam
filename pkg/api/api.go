package api

import (
	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var headers = map[string]string{"Access-Control-Allow-Origin": "*"} // Required when CORS enabled in API Gateway.

func Err(err error) (events.APIGatewayV2HTTPResponse, error) {
	return worker(http.StatusBadRequest, err.Error())
}

func OK(body string) (events.APIGatewayV2HTTPResponse, error) {
	return worker(http.StatusOK, body)
}

func worker(code int, body string) (events.APIGatewayV2HTTPResponse, error) {
	r := events.APIGatewayV2HTTPResponse{Headers: headers, StatusCode: code, Body: body}
	log.WithFields(log.Fields{"res": r}).Info()
	return r, nil
}
