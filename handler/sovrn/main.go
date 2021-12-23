package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/dustin/go-humanize"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"plumbus/pkg/api"
	"plumbus/pkg/util"
	"strconv"
	"strings"
	"time"
)

var (
	db        *dynamodb.Client
	tableName = "plumbus_fb_sovrn"
)

type Entity struct {

	// Campaign is the partition key.
	Campaign string `json:"campaign"`

	// Unix is the UTC time in seconds when this entity was created.
	Unix int64 `json:"unix"`

	AdSet     int     `json:"ad_set"`
	SubID     int     `json:"sub_id"`
	Revenue   float64 `json:"revenue"`
	Formatted string  `json:"formatted"`
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})
	if cfg, err := config.LoadDefaultConfig(context.TODO()); err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		panic(err)
	} else {
		db = dynamodb.NewFromConfig(cfg)
	}
}

func handle(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"req.Body": request.Body}).Info()

	var err error

	p := strings.Split(request.Body, "Impressions UTM Adset,Impressions UTM SubID,Impressions UTM Campaign,Impressions Estimated Revenue")
	if len(p) == 0 {
		err = errors.New("unable to split")
		return api.Err(err)
	}

	// now split by newline chars boom check ah wow
	heyNow := strings.Split(p[1], "\\n")

	var entities []Entity

	for _, hey := range heyNow {

		fields := strings.Split(hey, ",")

		if len(fields) != 4 {
			fmt.Println("wonky")
			continue
		}

		input := &dynamodb.PutItemInput{
			Item: map[string]types.AttributeValue{
				"ad_set":    &types.AttributeValueMemberN{Value: fields[0]},
				"sub_id":    &types.AttributeValueMemberN{Value: fields[1]},
				"campaign":  &types.AttributeValueMemberN{Value: fields[2]},
				"revenue":   &types.AttributeValueMemberN{Value: fields[3]},
				"formatted": &types.AttributeValueMemberS{Value: "$" + humanize.FormatFloat("#,###.##", util.StringToFloat64(fields[3]))},
				"unix":      &types.AttributeValueMemberN{Value: strconv.Itoa(int(time.Now().UTC().Unix()))},
			},
			TableName: &tableName,
		}

		out, err := db.PutItem(context.TODO(), input)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error()
			continue
		}

		var e Entity
		err = attributevalue.UnmarshalMap(out.Attributes, &e)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error()
			continue
		}

		//var item map[string]types.AttributeValue{}
		//if item, err = attributevalue.MarshalMap(&e); err != nil {
		//	fmt.Println("unable to marshal map")
		//	log.WithFields(log.Fields{"err": err}).Error()
		//	continue
		//}

		//if _, err := db.PutItem(context.TODO(), &dynamodb.PutItemInput{Item: item, TableName: &tableName}); err != nil {
		//	log.WithFields(log.Fields{"err": err}).Error()
		//	continue
		//}
		entities = append(entities, e)
	}

	var bytes []byte
	if bytes, err = json.Marshal(&entities); err != nil {
		return api.Err(err)
	}

	return api.OK(string(bytes))
}

func put(body string) (string, error) {

	var m map[string]interface{}
	_ = json.Unmarshal([]byte(body), &m)

	item, _ := attributevalue.MarshalMap(m)

	out, err := db.PutItem(context.TODO(), &dynamodb.PutItemInput{
		Item:      item,
		TableName: &tableName,
	})

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error()
		return "", err
	}

	_ = attributevalue.UnmarshalMap(out.Attributes, &m)

	bytes, _ := json.Marshal(m)

	return string(bytes), nil
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
