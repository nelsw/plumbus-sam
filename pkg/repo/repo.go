package repo

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/util/logs"
)

var (
	db  *dynamodb.Client
	ctx = context.TODO()
)

func init() {

	logs.Init()

	if cfg, err := config.LoadDefaultConfig(ctx); err != nil {
		log.WithError(err).Fatal()
	} else {
		db = dynamodb.NewFromConfig(cfg)
	}
}

func DelByEntry(table, key, val string) error {
	return DelByKey(table, map[string]types.AttributeValue{key: &types.AttributeValueMemberS{Value: val}})
}

func DelByKeys(table string, keys []map[string]types.AttributeValue) error {
	for _, key := range keys {
		if err := DelByKey(table, key); err != nil {
			log.WithError(err).Error()
			return err
		}
	}
	return nil
}

func DelByKey(table string, key map[string]types.AttributeValue) error {
	if _, err := db.DeleteItem(ctx, &dynamodb.DeleteItemInput{TableName: &table, Key: key}); err != nil {
		log.WithError(err).Error()
		return err
	}
	return nil
}

func ScanTable(table string) (*dynamodb.ScanOutput, error) {
	return ScanInput(&dynamodb.ScanInput{TableName: &table})
}

func ScanInputAndUnmarshal(input *dynamodb.ScanInput, out *[]interface{}) error {
	if output, err := ScanInput(input); err != nil {
		log.WithError(err).Error()
		return err
	} else if err = attributevalue.UnmarshalListOfMaps(output.Items, out); err != nil {
		log.WithError(err).Error()
		return err
	} else {
		return nil
	}
}

func ScanInput(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	if out, err := db.Scan(ctx, input); err != nil {
		log.WithError(err).Error()
		return nil, err
	} else {
		return out, nil
	}
}

func Put(input *dynamodb.PutItemInput) error {
	if _, err := db.PutItem(ctx, input); err != nil {
		log.WithError(err).Error()
		return err
	}
	return nil
}
