package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/fb"
	"plumbus/pkg/repo"
	"plumbus/pkg/util/logs"
)

func init() {
	logs.Init()
}

func handle(in map[string]interface{}) (map[string]interface{}, error) {

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

	if id, ok := in["account"]; ok {
		return account(fmt.Sprintf("%v", id), ignore)
	}

	return nil, errors.New("key not found")
}

func accounts(ignore map[string]interface{}) (map[string]interface{}, error) {
	if out, err := fb.Accounts().Marketing().GET(); err != nil {
		log.WithError(err).Error()
		return nil, err
	} else {
		res := map[string]interface{}{}
		for k, v := range out {
			if _, ok := ignore[k]; !ok {
				res[k] = v
			}
		}
		return res, nil
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
