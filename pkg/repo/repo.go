// Package repo provides common db functionality.
// DynamoDB reserved keywords: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/ReservedWords.html
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

const maxRequestSize = 25 // you can afford more than this jeff

var db *dynamodb.Client

func init() {
	logs.Init()
	if cfg, err := config.LoadDefaultConfig(context.Background()); err != nil {
		log.WithError(err).Fatal()
	} else {
		db = dynamodb.NewFromConfig(cfg)
	}
}

func Get(ctx context.Context, table, key, val string, v interface{}) error {

	input := &dynamodb.GetItemInput{
		TableName: &table,
		Key: map[string]types.AttributeValue{
			key: &types.AttributeValueMemberS{
				Value: val,
			},
		},
	}

	if output, err := db.GetItem(ctx, input); err != nil {
		return err
	} else if err = attributevalue.UnmarshalMap(output.Item, &v); err != nil {
		return err
	} else {
		return nil
	}
}

func Delete(ctx context.Context, in *dynamodb.DeleteItemInput) (err error) {
	_, err = db.DeleteItem(ctx, in)
	return
}

func Scan(ctx context.Context, in *dynamodb.ScanInput, v interface{}) error {
	if out, err := db.Scan(ctx, in); err != nil {
		return err
	} else {
		return attributevalue.UnmarshalListOfMaps(out.Items, &v)
	}
}

func Put(ctx context.Context, input *dynamodb.PutItemInput) error {
	_, err := db.PutItem(ctx, input)
	return err
}

func Update(ctx context.Context, in *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return db.UpdateItem(ctx, in)
}

func BatchWrite(ctx context.Context, table string, rr []types.WriteRequest) (err error) {

	var in *dynamodb.BatchWriteItemInput
	for _, r := range chunkWriteRequests(rr) {
		in = &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				table: r,
			},
		}
		if _, err = db.BatchWriteItem(ctx, in); err != nil {
			break
		}
	}

	return
}

func BatchGet(ctx context.Context, in *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error) {
	return db.BatchGetItem(ctx, in)
}

func Query(ctx context.Context, in *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return db.Query(ctx, in)
}

func chunkWriteRequests(in []types.WriteRequest) (out [][]types.WriteRequest) {
	var end int
	for i := 0; i < len(in); i += maxRequestSize {
		if end = i + maxRequestSize; end > len(in) {
			end = len(in)
		}
		out = append(out, in[i:end])
	}
	return
}
