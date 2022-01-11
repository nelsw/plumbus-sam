package fb

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/fb"
	"plumbus/pkg/repo"
	"plumbus/pkg/util"
	"strings"
	"time"
)

const (
	api = "https://graph.facebook.com/v12.0"
)

type Payload struct {
	Data []interface{} `json:"data"`
	Page struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type Account struct {
	ID        string     `json:"id"`
	AccountID string     `json:"account_id"`
	Name      string     `json:"name"`
	Status    int        `json:"account_status"`
	Created   string     `json:"created_time"`
	Campaigns []Campaign `json:"children,omitempty"`
}

type Campaign struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	DailyBudget     string `json:"daily_budget,omitempty"`
	RemainingBudget string `json:"remaining_budget,omitempty"`
	Created         string `json:"created_time"`
	Updated         string `json:"updated_time"`
}

func (c Campaign) UTM() string {
	utm := getParenWrappedSovrnID(c.ID)
	if utm == "" {
		if spaced := strings.Split(c.Name, " "); len(spaced) > 1 && util.IsNumber(spaced[0]) {
			utm = spaced[0]
		} else if scored := strings.Split(c.Name, "_"); len(scored) > 1 && util.IsNumber(scored[0]) {
			utm = scored[0]
		} else {
			utm = c.ID
		}
	}
	return utm
}

func getParenWrappedSovrnID(s string) string {
	if chunks := strings.Split(s, "("); len(chunks) > 1 {
		chunk := chunks[1]
		chunks = strings.Split(chunk, ")")
		if len(chunks) > 1 {
			return chunks[0]
		}
	}
	return ""
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

	var payload fb.Payload
	if err = json.Unmarshal(body, &payload); err != nil {
		log.WithError(err).Error()
		return
	}

	if data = append(data, payload.Data...); payload.Page.Next == "" {
		return
	}

	var next []interface{}
	if next, err = Get(payload.Page.Next); err != nil {
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
