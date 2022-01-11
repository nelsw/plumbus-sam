package api

import (
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

func OK(body string) (events.APIGatewayV2HTTPResponse, error) {
	return worker(http.StatusOK, body)
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
