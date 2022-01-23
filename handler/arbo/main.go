package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/model/arbo"
	"plumbus/pkg/repo"
	"plumbus/pkg/util/logs"
)

var client = &http.Client{}

func init() {
	logs.Init()
}

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.WithFields(log.Fields{"ctx": ctx, "reg": req}).Info()
	switch req.RequestContext.HTTP.Method {
	case http.MethodOptions:
		return api.K()
	case http.MethodGet:
		return get(ctx, req)
	case http.MethodPut:
		return put(ctx)
	default:
		return api.Nada()
	}
}

func get(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if id, ok := req.QueryStringParameters["id"]; ok {
		return batch(ctx, id)
	} else {
		return scan(ctx)
	}
}

func batch(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {

	in := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			arbo.Table(): {
				Keys: []map[string]types.AttributeValue{{"ID": &types.AttributeValueMemberS{Value: id}}},
			},
		},
	}

	var err error
	var out *dynamodb.BatchGetItemOutput
	if out, err = repo.BatchGet(ctx, in); err != nil {
		log.WithError(err).Error()
		return api.Err(err)
	}

	var arr []arbo.Entity
	if err = attributevalue.UnmarshalListOfMaps(out.Responses[arbo.Table()], &arr); err != nil {
		log.WithError(err).Error()
		return api.Err(err)
	}

	log.Trace("batch get for arbo campaign ", id, " found ", len(arr))
	return api.JSON(arr)
}

func scan(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	var out []arbo.Entity
	if err := repo.Scan(ctx, &dynamodb.ScanInput{TableName: arbo.TableName()}, &out); err != nil {
		log.WithError(err).Error("scanning arbo table")
		return api.Err(err)
	}
	return api.JSON(out)
}

func put(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	var ee []arbo.Entity

	for _, c := range arbo.Clients() {
		if arr, err := fetch(ctx, c); err != nil {
			log.WithError(err).Error("fetch ", c)
			return api.Err(err)
		} else {
			ee = append(ee, arr...)
		}
	}

	log.Info("total arbo entities fetched: ", len(ee))
	// 4,027

	var rr []types.WriteRequest
	for _, e := range ee {
		rr = append(rr, e.WriteRequest())
	}

	if err := repo.BatchWrite(ctx, arbo.Table(), rr); err != nil {
		log.WithError(err).Error("writing arbo data")
		return api.Err(err)
	}

	return api.JSON(ee)
}

func fetch(ctx context.Context, c arbo.Client) ([]arbo.Entity, error) {

	log.Trace("fetching ", c)

	var err error

	var res *http.Response
	if res, err = client.Do(arbo.NewRequest(ctx, c)); err != nil {
		log.WithError(err).Error("campaign request failed for ", c)
		return nil, err
	}

	log.Trace("campaign request response status code ", res.StatusCode)

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			log.WithError(err).Error("error closing body")
		}
	}(res.Body)

	var zr *gzip.Reader
	if zr, err = gzip.NewReader(res.Body); err != nil {
		log.WithError(err).Error("gzip reader failed")
		return nil, err
	}

	var out bytes.Buffer
	var wrt int64
	if wrt, err = io.Copy(&out, zr); err != nil {
		log.WithError(err).Error("writing failed")
		return nil, err
	}

	log.Trace("bytes copied: ", wrt)

	var pay arbo.Payload
	if err = json.Unmarshal(out.Bytes(), &pay); err != nil {
		log.WithError(err).Error("error unmarshalling into payload")
		return nil, err
	}

	log.Trace("fetched entities: ", len(pay.Data))

	return pay.Data, nil
}

func main() {
	lambda.Start(handle)
}
