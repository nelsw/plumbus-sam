package repo

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"strings"
)

const maxRequestSize = 25 // you can afford more than this jeff

var db *dynamodb.Client

func init() {
	if cfg, err := config.LoadDefaultConfig(context.Background()); err != nil {
		log.Panic(err)
	} else {
		db = dynamodb.NewFromConfig(cfg)
	}
}

func DeleteItem(ctx context.Context, in *dynamodb.DeleteItemInput) (err error) {
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

func BatchWriteItems(ctx context.Context, table string, rr []types.WriteRequest) error {

	var ee []error

	var in *dynamodb.BatchWriteItemInput
	for _, r := range chunkWriteRequests(rr) {
		in = &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				table: r,
			},
		}
		if _, err := db.BatchWriteItem(ctx, in); err != nil {
			ee = append(ee, err)
		}
	}

	if len(ee) == 0 {
		return nil
	}

	var ss []string
	for _, e := range ee {
		ss = append(ss, e.Error())
	}

	return errors.New(strings.Join(ss, "\n"))
}

func BatchGetItem(ctx context.Context, in *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error) {
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
