package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	fass_types "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/api"
	"plumbus/pkg/model/fb"
	"plumbus/pkg/repo"
	"plumbus/pkg/util"
	"plumbus/pkg/util/logs"
	"sync"
)

const (
	v12 = "https://graph.facebook.com/v12.0"
)

var (
	sam   *faas.Client
	mutex = sync.Mutex{}
)

func init() {
	logs.Init()
	if cfg, err := config.LoadDefaultConfig(context.TODO()); err != nil {
		log.WithError(err).Fatal()
	} else {
		sam = faas.NewFromConfig(cfg)
	}
}

func handle(ctx context.Context, event events.CloudWatchEvent) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "event": event}).Info()

	var err error

	var accounts []fb.Account
	if accounts, err = tree(ctx); err != nil {
		return api.Err(err)
	}

	agg := map[string]fb.Campaign{}

	var chunk []map[string]types.AttributeValue

	var fin []Sovrn

	var wg1 sync.WaitGroup
	for index, account := range accounts {

		if len(account.Campaigns) == 0 {
			continue
		}

		wg1.Add(1)

		go func(index int, max int, account fb.Account) {

			defer wg1.Done()

			for _, campaign := range account.Campaigns {

				if campaign.UTM() == "" {
					continue
				}

				mutex.Lock()
				if _, ok := agg[campaign.UTM()]; ok {
					mutex.Unlock()
					continue
				}

				agg[campaign.UTM()] = campaign
				mutex.Unlock()

				if len(chunk) > 0 && len(chunk)%25 == 0 {
					var got []Sovrn
					if got, err = batchGet(ctx, chunk); err != nil {
						log.WithError(err).Error()
					} else if len(got) > 0 {

						for _, g := range got {

							mutex.Lock()
							d := agg[g.Campaign]
							mutex.Unlock()

							g.Detail = d

							var all []interface{}
							if all, err = get(campaignsUrl(d.ID)); err != nil {
								log.WithError(err).Error()
							} else if len(all) > 0 {
								a := all[0]
								g.Spend = a.(map[string]interface{})["spend"].(string)
							}

							fin = append(fin, g)
						}
					}
					chunk = []map[string]types.AttributeValue{}
				}

				chunk = append(chunk, map[string]types.AttributeValue{"campaign": &types.AttributeValueMemberS{Value: campaign.UTM()}})
			}

			if len(chunk) > 0 && (index+1 == max || len(chunk)%25 == 0) {
				var got []Sovrn
				if got, err = batchGet(ctx, chunk); err != nil {
					log.WithError(err).Error()
				} else if len(got) > 0 {
					for _, g := range got {

						mutex.Lock()
						d := agg[g.Campaign]
						mutex.Unlock()

						g.Detail = d

						var all []interface{}
						if all, err = get(campaignsUrl(d.ID)); err != nil {
							log.WithError(err).Error()
						} else if len(all) > 0 {
							a := all[0]
							g.Spend = a.(map[string]interface{})["spend"].(string)
						}
						fin = append(fin, g)
					}
				}
			}

		}(index, len(accounts), account)
	}

	wg1.Wait()

	util.PrettyPrint(fin)

	return api.OK("")
}

type Sovrn struct {
	Campaign    string      `json:"campaign"`
	Revenue     float64     `json:"revenue"`
	CTR         float64     `json:"ctr"`
	Impressions int         `json:"impressions"`
	Sessions    int         `json:"sessions"`
	PageViews   int         `json:"page_views"`
	Detail      fb.Campaign `json:"detail"`
	Spend       string      `json:"spend"`
}

func batchGet(ctx context.Context, keys []map[string]types.AttributeValue) (got []Sovrn, err error) {
	in := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			"plumbus_fb_sovrn": {
				Keys: keys,
			},
		},
	}
	var out *dynamodb.BatchGetItemOutput
	if out, err = repo.BatchGetItem(ctx, in); err != nil {
		log.WithError(err).Error()
	} else if err = attributevalue.UnmarshalListOfMaps(out.Responses["plumbus_fb_sovrn"], &got); err != nil {
		log.WithError(err).Error()
	}
	return
}

func tree(ctx context.Context) ([]fb.Account, error) {

	in := &faas.InvokeInput{
		FunctionName:   ptr.String("plumbus_treeHandler"),
		InvocationType: fass_types.InvocationTypeRequestResponse,
		LogType:        fass_types.LogTypeTail,
	}

	var err error

	var out *faas.InvokeOutput
	if out, err = sam.Invoke(ctx, in); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var response events.APIGatewayV2HTTPResponse
	if err = json.Unmarshal(out.Payload, &response); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var body string
	if body = response.Body; response.StatusCode != http.StatusOK {
		err = errors.New(body)
		log.WithError(err).Error()
		return nil, err
	}

	var accounts []fb.Account
	if err = json.Unmarshal([]byte(body), &accounts); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	return accounts, nil
}

func get(url string) (data []interface{}, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {
		log.WithError(err).Error()
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		log.WithError(err).Error()
		return
	}

	var payload fb.Payload
	if err = json.Unmarshal(body, &payload); err != nil {
		log.WithError(err).Error()
		return
	}

	if data = append(data, payload.Data...); payload.Page.Next == "" {
		return
	}

	var next []interface{}
	if next, err = get(payload.Page.Next); err != nil {
		log.WithError(err).Error()
		return
	}

	data = append(data, next...)
	return
}

func token() string {
	return "?access_token=" + os.Getenv("tkn")
}

func campaignsUrl(ID string) string {
	fields := "&fields=spend"
	date := "&date_preset=today"
	return v12 + "/" + ID + "/insights" + token() + fields + date
}

func main() {
	lambda.Start(handle)
}
