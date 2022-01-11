package repo

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/util/logs"
	"time"
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

func GetIt(table, key, val string, it interface{}) (err error) {

	var data []byte
	if data, err = get(table, key, val, 6); err != nil {
		log.WithError(err).Error()
	} else if err = json.Unmarshal(data, &it); err != nil {
		log.WithError(err).Error()
	}

	return
}

func Get(table, key, val string) ([]byte, error) {
	return get(table, key, val, 1)
}

func get(table, key, val string, attempt int64) ([]byte, error) {
	var err error

	var output *dynamodb.GetItemOutput
	var input = &dynamodb.GetItemInput{
		TableName: &table,
		Key:       map[string]types.AttributeValue{key: &types.AttributeValueMemberS{Value: val}},
	}

	if output, err = db.GetItem(ctx, input); err != nil {
		if attempt > 5 {
			log.WithError(err).Error()
			return nil, err
		}
		log.Trace(err)
		time.Sleep(time.Duration(1000 * 2 * attempt))
		return get(table, key, val, attempt+1)
	}

	var payload map[string]interface{}
	if err = attributevalue.UnmarshalMap(output.Item, &payload); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var bytes []byte
	if bytes, err = json.Marshal(&payload); err != nil {
		return nil, err
	}

	return bytes, nil
}

func Exists(table, key, val string) (bool, error) {

	var err error
	var output *dynamodb.GetItemOutput
	var input = &dynamodb.GetItemInput{
		TableName: &table,
		Key:       map[string]types.AttributeValue{key: &types.AttributeValueMemberS{Value: val}},
	}

	if output, err = db.GetItem(ctx, input); err != nil {
		log.WithError(err).Error()
		return false, err
	}

	var payload map[string]interface{}
	if err = attributevalue.UnmarshalMap(output.Item, &payload); err != nil {
		log.WithError(err).Error()
		return false, err
	}

	_, exists := payload[key]
	return exists, nil
}

func DelByEntry(table, key, val string) error {
	return DelByKey(table, map[string]types.AttributeValue{key: &types.AttributeValueMemberS{Value: val}})
}

func DelByKey(table string, key map[string]types.AttributeValue) error {
	if _, err := db.DeleteItem(ctx, &dynamodb.DeleteItemInput{TableName: &table, Key: key}); err != nil {
		log.WithError(err).Error()
		return err
	}
	return nil
}

func ScanInputAndUnmarshal(input *dynamodb.ScanInput, out *interface{}) error {
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

func Scan(in *dynamodb.ScanInput, v interface{}) (err error) {

	var out *dynamodb.ScanOutput
	if out, err = db.Scan(ctx, in); err != nil {
		log.WithError(err).Error()
		return
	}

	if err = attributevalue.UnmarshalListOfMaps(out.Items, &v); err != nil {
		log.WithError(err).Error()
		return err
	}

	return
}

func Put(input *dynamodb.PutItemInput) error {
	if _, err := db.PutItem(ctx, input); err != nil {
		log.WithError(err).Error()
		return err
	}
	return nil
}

func BatchWriteItem(ctx context.Context, in *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
	return db.BatchWriteItem(ctx, in)
}

func BatchWriteItems(ctx context.Context, table string, requests []types.WriteRequest) (*dynamodb.BatchWriteItemOutput, error) {
	in := &dynamodb.BatchWriteItemInput{RequestItems: map[string][]types.WriteRequest{table: requests}}
	return BatchWriteItem(ctx, in)
}

func BatchGetItem(ctx context.Context, in *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error) {
	return db.BatchGetItem(ctx, in)
}

func Query(ctx context.Context, in *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return db.Query(ctx, in)
}
