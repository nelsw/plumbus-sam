package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/fb"
	"plumbus/pkg/repo"
	"plumbus/pkg/util"
	"plumbus/pkg/util/logs"
	"sync"
)

var mutex = &sync.Mutex{}

func init() {
	logs.Init()
}

func handle(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {

	log.Info(in)

	ignore, err := accountsToIgnore()

	if err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	if _, ok := in["ignore"]; ok {
		return ignore, nil
	}

	if _, ok := in["accounts"]; ok {
		return accounts(ignore)
	}

	if id, ok := in["campaigns"]; ok {
		return campaigns(fmt.Sprintf("%v", id))
	}

	if id, ok := in["account"]; ok {
		return account(fmt.Sprintf("%v", id), ignore)
	}

	return nil, errors.New("key not found")
}

func accounts(ignore map[string]interface{}) (map[string]interface{}, error) {

	if out, err := fb.Accounts(ignore).Marketing().GET(); err != nil {
		log.WithError(err).Error()
		return nil, err
	} else {

		result := map[string]interface{}{
			"total_spend":   float64(0),
			"total_revenue": float64(0),
			"total_profit":  float64(0),
		}

		var wg sync.WaitGroup
		for k, v := range out {

			wg.Add(1)

			go func(k string, v interface{}) {
				defer wg.Done()
				var res map[string]interface{}
				if res, err = campaigns(k); err != nil {
					log.WithError(err)
				} else {
					spend := util.StringToFloat64(fmt.Sprintf("%v", res["account_spend"]))
					revenue := util.StringToFloat64(fmt.Sprintf("%v", res["account_revenue"]))
					profit := revenue - spend

					mutex.Lock()
					result["total_spend"] = result["total_spend"].(float64) + spend
					result["total_revenue"] = result["total_revenue"].(float64) + revenue
					result["total_profit"] = result["total_profit"].(float64) + profit
					result[k] = map[string]interface{}{
						"account":   v,
						"campaigns": res,
					}
					mutex.Unlock()
				}
			}(k, v)
		}
		wg.Wait()
		return result, nil
	}
}

func campaigns(accountID string) (map[string]interface{}, error) {
	if out, err := fb.Campaigns(accountID).Insights().GET(); err != nil {
		log.WithError(err).Error()
		return nil, err
	} else {
		return out, nil
	}
}

func account(accountID string, ignore map[string]interface{}) (map[string]interface{}, error) {
	if _, found := ignore[accountID]; found {
		return nil, errors.New("requested account is ignored")
	}
	if out, err := fb.Account(accountID).Marketing().GET(); err != nil {
		log.WithError(err).Error()
		return nil, err
	} else {
		return out, nil
	}
}

func accountsToIgnore() (map[string]interface{}, error) {
	in := &dynamodb.ScanInput{TableName: ptr.String("plumbus_ignored_ad_accounts")}
	var out interface{}
	if err := repo.ScanInputAndUnmarshal(in, &out); err != nil {
		log.WithError(err).Error()
		return nil, err
	}
	res := map[string]interface{}{}
	for _, w := range out.([]interface{}) {
		z := w.(map[string]interface{})
		res[fmt.Sprintf("%v", z["account_id"])] = true
	}
	return res, nil
}

func main() {
	lambda.Start(handle)
}
