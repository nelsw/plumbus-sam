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

type Payload struct {
	Attachment struct {
		Data string `json:"data"`
	} `json:"attachment"`
}

var (
	db        *dynamodb.Client
	tableName = "plumbus_fb_rev"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	if cfg, err := config.LoadDefaultConfig(context.TODO()); err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		panic(err)
	} else {
		db = dynamodb.NewFromConfig(cfg)
	}
}

func handle(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	//for key := range accountsToSkip {
	//	_, err := db.PutItem(context.TODO(), &dynamodb.PutItemInput{
	//		Item:      map[string]types.AttributeValue{"account_id": &types.AttributeValueMemberS{Value: key}},
	//		TableName: aws.String("plumbus_ignored_ad_accounts"),
	//	})
	//	if err != nil {
	//		fmt.Println(key, err)
	//	}
	//}

	//log.WithFields(log.Fields{"request": request}).Info()
	//
	//fmt.Println(request.Body)
	//
	//var p Payload
	//
	//var err error
	//
	//err = json.Unmarshal([]byte(request.Body), &p)
	//
	//log.WithFields(log.Fields{"payload": p}).Info()
	//
	//sDec, _ := b64.StdEncoding.DecodeString(p.Attachment.Data)
	//fmt.Println(string(sDec))
	//fmt.Println()
	//
	//var reader *zip.Reader
	//if reader, err = zip.NewReader(bytes.NewReader(sDec), int64(len(sDec))); err != nil {
	//	fmt.Println("unable to read")
	//	fmt.Println(err)
	//	return response(http.StatusBadRequest)
	//}
	//
	//for _, file := range reader.File {
	//
	//	if file.Name != "dashboard-acquisition/performance.csv" {
	//		continue
	//	}
	//
	//	var closer io.ReadCloser
	//	if closer, err = file.Open(); err != nil {
	//		return response(http.StatusBadRequest)
	//	}
	//
	//	var data [][]string
	//	if data, err = csv.NewReader(closer).ReadAll(); err != nil {
	//		return response(http.StatusBadRequest)
	//	}
	//
	//	fmt.Println(data)
	//}

	return response(http.StatusOK)
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

func response(code int) (events.APIGatewayV2HTTPResponse, error) {
	r := events.APIGatewayV2HTTPResponse{
		Headers:    map[string]string{"Access-Control-Allow-Origin": "*"}, // Required when CORS enabled in API Gateway.
		StatusCode: code,
	}
	log.WithFields(log.Fields{"response": r}).Info()
	return r, nil
}

func main() {
	lambda.Start(handle)
}
