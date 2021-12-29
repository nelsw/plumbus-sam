package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	db     *dynamodb.Client
	ctx    = context.TODO()
	params = &dynamodb.ScanInput{TableName: aws.String("plumbus_ignored_ad_accounts")}
)

func init() {

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
		ForceColors:   true,
	})

	if cfg, err := config.LoadDefaultConfig(ctx); err != nil {
		log.WithError(err).Fatal()
	} else {
		db = dynamodb.NewFromConfig(cfg)
	}
}

func handle() (out []interface{}, err error) {

	var scanOutput *dynamodb.ScanOutput
	if scanOutput, err = db.Scan(ctx, params); err == nil {
		if err = attributevalue.UnmarshalListOfMaps(scanOutput.Items, &out); err == nil {
			return
		}
	}

	log.WithError(err).Error()
	return
}

func main() {
	lambda.Start(handle)
}
