package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

var (
	db        *dynamodb.Client
	tableName = "plumbus_fb"
)

func init() {
	log.SetOutput(os.Stdout)
	if cfg, err := config.LoadDefaultConfig(context.TODO()); err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		panic(err)
	} else {
		db = dynamodb.NewFromConfig(cfg)
	}
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

func put(body string) (string, int) {

	var m map[string]interface{}
	_ = json.Unmarshal([]byte(body), &m)

	item, _ := attributevalue.MarshalMap(m)

	out, err := db.PutItem(context.TODO(), &dynamodb.PutItemInput{
		Item:      item,
		TableName: &tableName,
	})

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		return "", http.StatusBadRequest
	}

	_ = attributevalue.UnmarshalMap(out.Attributes, &m)

	bytes, _ := json.Marshal(m)

	return string(bytes), http.StatusOK
}

func get(ids []string) (string, int) {

	if ids == nil {
		return "", http.StatusBadRequest
	}

	if len(ids) == 1 {

		out, err := db.GetItem(context.TODO(), &dynamodb.GetItemInput{
			Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: ids[0]}},
			TableName: &tableName,
		})
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error()
			return "", http.StatusBadRequest
		}

		var m map[string]interface{}
		_ = attributevalue.UnmarshalMap(out.Item, &m)
		bytes, _ := json.Marshal(m)

		if len(m) == 0 {
			return "", http.StatusNotFound
		}

		return string(bytes), http.StatusOK
	}

	var keys []map[string]types.AttributeValue
	for _, id := range ids {
		keys = append(keys, map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: id}})
	}

	out, err := db.BatchGetItem(context.TODO(), &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{tableName: {Keys: keys}},
	})
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		return "", http.StatusBadRequest
	}

	var l []map[string]interface{}
	_ = attributevalue.UnmarshalListOfMaps(out.Responses[tableName], &l)
	bytes, _ := json.Marshal(l)

	if len(l) == 0 {
		return "", http.StatusNotFound
	}

	log.WithFields(log.Fields{"l": l}).Info("l")

	return string(bytes), http.StatusOK
}

func main() {
	lambda.Start(handle)
}
