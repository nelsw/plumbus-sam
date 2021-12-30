package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/fb"
	"plumbus/pkg/repo"
	"plumbus/pkg/util/logs"
)

var (
	key   = "account_id"
	table = "plumbus_ignored_ad_accounts"
	input = &dynamodb.ScanInput{TableName: &table}
)

func init() {
	logs.Init()
}

func handle(ctx context.Context) (out map[string]interface{}, err error) {

	var arr interface{}
	if arr, err = getAccountsToIgnore(); err != nil {
		log.WithError(err).Error()
		return
	}

	var bytes []byte
	if bytes, err = json.Marshal(&arr); err != nil {
		log.WithError(err).Error()
		return
	}

	var bads []struct {
		AccountID string `json:"account_id"`
	}
	if err = json.Unmarshal(bytes, &bads); err != nil {
		log.WithError(err).Error()
		return
	}

	var value interface{}

	/*
		return info for all valid accounts
	*/
	if value = ctx.Value("accounts"); value != nil {

		if out, err = fb.Accounts().Marketing().GET(); err != nil {
			log.WithError(err).Error()
			return
		}

		bad := map[string]interface{}{}
		for _, b := range bads {
			bad[b.AccountID] = b.AccountID
		}

		for k := range out {
			if _, ok := bad[k]; ok {
				delete(out, k)
			} else {
				var tmp map[string]interface{}
				if tmp, err = getAccountById(k); err != nil {
					log.WithError(err).Error()
				} else {
					out[k] = tmp
				}
			}
		}

		return
	}

	/*
		return info for a single account_id
	*/
	if value = ctx.Value(key); value != nil {
		if out, err = getAccountById(value); err != nil {
			log.WithError(err).Error()
		}
		return
	}

	/*
		return info regarding accounts to ignore
	*/
	out = map[string]interface{}{}
	for _, b := range bads {
		out[b.AccountID] = true
	}

	return
}

func getAccountById(value interface{}) (out map[string]interface{}, err error) {

	var accountID = fmt.Sprintf("%v", value)

	var exists bool
	if exists, err = repo.Exists(table, key, accountID); err != nil {
		log.WithError(err).Error()
		return
	}

	if exists {
		return
	}

	if out, err = fb.Account(accountID).Marketing().GET(); err != nil {
		log.WithError(err).Error()
	}

	return
}

func getAccountsToIgnore() (arr interface{}, err error) {
	if err = repo.ScanInputAndUnmarshal(input, &arr); err != nil {
		log.WithError(err).Error()
	}
	return
}

func main() {
	lambda.Start(handle)
}
