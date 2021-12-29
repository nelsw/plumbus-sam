package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
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

	accountsToSkip := map[string]interface{}{
		"413679809716807":   nil,
		"750877365817118":   nil,
		"192357189581498":   nil,
		"612031486585483":   nil,
		"3575472439245925":  nil,
		"619944848765264":   nil,
		"218941270288997":   nil,
		"245495866953965":   nil,
		"5688782297830054":  nil,
		"491184258440143":   nil,
		"2403739559944547":  nil,
		"12817637":          nil,
		"3671868129555357":  nil,
		"333485221856766":   nil,
		"155263960129351":   nil,
		"2177908659168303":  nil,
		"625915844866526":   nil,
		"2180764178898025":  nil,
		"246088729880368":   nil,
		"1097225754361421":  nil,
		"1243882536086716":  nil,
		"254126193351151":   nil,
		"931887524201557":   nil,
		"428801301949781":   nil,
		"913678565938954":   nil,
		"1515798122112809":  nil,
		"546285956018323":   nil,
		"259223871759580":   nil,
		"547586279498916":   nil,
		"209494886979585":   nil,
		"1450566098533975":  nil,
		"426011485278946":   nil,
		"1171166016595742":  nil,
		"830846364118052":   nil,
		"788842328547187":   nil,
		"222205359142916":   nil,
		"1408104866244851":  nil,
		"338022787312593":   nil,
		"665195921058997":   nil,
		"2326888560975233":  nil,
		"10150903730580756": nil,
		"302191798223982":   nil,
		"248532193069417":   nil,
		"279967990799943":   nil,
	}

	for key := range accountsToSkip {
		_, err := db.PutItem(context.TODO(), &dynamodb.PutItemInput{
			Item:      map[string]types.AttributeValue{"account_id": &types.AttributeValueMemberS{Value: key}},
			TableName: aws.String("plumbus_ignored_ad_accounts"),
		})
		if err != nil {
			fmt.Println(key, err)
		}
	}

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
