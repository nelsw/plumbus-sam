package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func handle(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"request": request}).Info()

	var body string
	code := http.StatusBadRequest

	response := events.APIGatewayV2HTTPResponse{
		Headers:    map[string]string{"Access-Control-Allow-Origin": "*"}, // Required when CORS enabled in API Gateway.
		StatusCode: code,
		Body:       body,
	}

	log.WithFields(log.Fields{"response": response}).Info()

	return response, nil
}

func main() {
	lambda.Start(handle)
}