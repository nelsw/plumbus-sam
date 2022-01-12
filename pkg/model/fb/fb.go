package fb

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/repo"
	"strings"
	"time"
)

const (
	api = "https://graph.facebook.com/v12.0"
)

type payload struct {
	Data []interface{} `json:"data"`
	Page struct {
		Next string `json:"next"`
	} `json:"paging"`
}

func Get(url string) (data []interface{}, err error) {
	return getAttempt(url, 1)
}

func getAttempt(url string, attempt int) (data []interface{}, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {

		if attempt > 9 {
			log.WithError(err).Error()
			return
		}

		if !strings.Contains(err.Error(), "too many open files") &&
			!strings.Contains(err.Error(), "no such host") {
			log.WithError(err).Trace()
		} else if strings.Contains(err.Error(), "connection refused") {
			log.WithError(err).Warn()
		}

		time.Sleep(time.Second * time.Duration(attempt))

		return getAttempt(url, attempt+1)
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		log.WithError(err).Error()
		return
	}

	var p payload
	if err = json.Unmarshal(body, &p); err != nil {
		log.WithError(err).Error()
		return
	}

	if data = append(data, p.Data...); p.Page.Next == "" {
		return
	}

	var next []interface{}
	if next, err = Get(p.Page.Next); err != nil {
		log.WithError(err).Error()
		return
	}

	data = append(data, next...)
	return
}

func Token() string {
	return "?access_token=" + os.Getenv("tkn")
}

func User() string {
	return os.Getenv("usr")
}

func API() string {
	return api
}

func AccountsToIgnore() (map[string]interface{}, error) {

	var in = dynamodb.ScanInput{TableName: ptr.String("plumbus_ignored_ad_accounts")}
	var out interface{}
	if err := repo.ScanInputAndUnmarshal(&in, &out); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	res := map[string]interface{}{}
	for _, w := range out.([]interface{}) {
		res[w.(map[string]interface{})["account_id"].(string)] = true
	}

	return res, nil
}
