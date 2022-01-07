package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/repo"
	"plumbus/pkg/util/logs"
)

var (
	input = &dynamodb.ScanInput{TableName: aws.String("plumbus_ignored_ad_accounts")}
)

func init() {
	logs.Init()
}

func handle() (out interface{}, err error) {
	if err := repo.ScanInputAndUnmarshal(input, &out); err != nil {
		log.WithError(err).Error()
	}
	return
}

func accountsToIgnore() (map[string]interface{}, error) {

	var in = dynamodb.ScanInput{TableName: ptr.String("plumbus_ignored_ad_accounts")}
	var out interface{}
	if err := repo.ScanInputAndUnmarshal(&in, &out); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	res := map[string]interface{}{}
	for _, w := range out.([]interface{}) {
		res[w.(map[string]interface{})["account_id"].(string)] = true
	}

	return res, nil
}

func main() {
	lambda.Start(handle)
}
