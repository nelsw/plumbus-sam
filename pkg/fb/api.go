package fb

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/objx"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/util"
	"plumbus/pkg/util/logs"
	"strconv"
	"strings"
)

var (
	accountMarketingFields = []string{
		"account_id",
		"account_status",
		"age",
		"amount_spent",
		"balance",
		"business",
		"created_time",
		"currency",
		"line_numbers",
		"media_agency",
		"name",
		"min_daily_budget",
		"owner",
		"spend_cap",
		"timezone_name",
		"timezone_offset_hours_utc",
	}
	campaignMarketingFields = []string{
		"id",
		"budget_remaining",
		"created_time",
		"name",
		"start_time",
		"status",
		"stop_time",
		"updated_time",
	}
	accountStatuses = map[int]string{
		1:   "Active",
		2:   "Disabled",
		3:   "Unsettled",
		7:   "PendingRiskReview",
		8:   "PendingSettlement",
		9:   "InGracePeriod",
		100: "PendingClosure",
		101: "Closed",
		201: "AnyActive",
		202: "AnyClosed",
	}
)

type API struct {
	account,
	campaign,
	marketing,
	insights bool

	accountID,
	campaignID,
	adSetID,
	adID,
	url string

	fields []string
}

func init() {
	logs.Init()
}

func Accounts() *API {
	a := new(API)
	a.account = true
	return a
}

func Account(id string) *API {
	a := new(API)
	a.account = true
	a.accountID = id
	return a
}

func Campaign(id string) *API {
	a := new(API)
	a.campaignID = id
	a.campaign = true
	return a
}

func Campaigns(id string) *API {
	a := new(API)
	a.accountID = id
	a.campaign = true
	return a
}

func AdSet(id string) *API {
	a := new(API)
	a.adSetID = id
	return a
}

func Ad(id string) *API {
	a := new(API)
	a.adID = id
	return a
}

func (a *API) Field(field string) *API {
	a.fields = append(a.fields, field)
	return a
}

func (a *API) Fields(fields []string) *API {
	a.fields = append(a.fields, fields...)
	return a
}

func (a *API) Insights() *API {
	a.insights = true
	return a
}

func (a *API) Marketing() *API {
	a.marketing = true
	if a.account {
		return a.Fields(accountMarketingFields)
	} else if a.campaign {
		return a.Fields(campaignMarketingFields)
	} else {
		return a
	}
}

func (a *API) GET() (map[string]interface{}, error) {

	a.setURL()

	if a.account && a.accountID == "" {
		return a.getAdAccounts()
	}

	if a.accountID != "" {

		a.setAdAccount()

		if a.campaign {

			a.setCampaigns()
			a.setToken()
			a.setFields()

			var err error
			var out []interface{}
			if out, err = get(a.url); err != nil {
				log.WithError(err).Error()
				return nil, err
			}

			var got = map[string]interface{}{}
			for _, o := range out {
				got[objx.New(o).Get("id").String()] = o
			}

			return got, nil

		}
	} else if a.campaignID != "" {
		a.url += a.campaignID
	} else if a.adSetID != "" {
		a.url += a.adSetID
	} else {
		a.url += a.adID
	}

	a.setToken()
	a.setFields()

	var err error
	//var out []interface{}
	//if out, err = get(url); err != nil {
	//	log.WithError(err).Error()
	//	return nil, err
	//}
	//
	//util.PrettyPrint(out)
	//
	//var got = map[string]interface{}{}
	//got["data"] = out

	var res *http.Response
	if res, err = http.Get(a.url); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var got = map[string]interface{}{}
	if err = json.Unmarshal(body, &got); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	if a.accountID != "" {
		formatAccountData(got)
	}

	return got, nil
}

func (a *API) setURL() {
	a.url = "https://graph.facebook.com/v12.0"
}

func (a *API) setUserID() {
	a.url += "/10158615602243295"
}

func (a *API) setAdAccounts() {
	a.url += "/adaccounts"
}

func (a *API) setAdAccount() {
	a.url += "/act_" + a.accountID
}

func (a *API) setCampaigns() {
	a.url += "/campaigns"
}

func (a *API) setToken() {
	a.url += "?access_token=" + os.Getenv("tkn")
}

func (a *API) setFields() {
	if len(a.fields) > 0 {
		a.url += "&fields=" + strings.Join(a.fields, ",")
	}
}

func (a *API) getAdAccounts() (map[string]interface{}, error) {

	a.setUserID()
	a.setAdAccounts()
	a.setToken()
	a.setFields()

	fmt.Println(a.url)

	var err error
	var out []interface{}
	if out, err = get(a.url); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var bytes []byte
	if bytes, err = json.Marshal(&out); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var accounts []struct {
		AccountID string `json:"account_id"`
	}

	if err = json.Unmarshal(bytes, &accounts); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	var got = map[string]interface{}{}
	for _, account := range accounts {
		got[account.AccountID] = nil
	}
	return got, nil
}

func formatAccountData(got map[string]interface{}) {

	/*
		human-readable account status
	*/
	if ko, ok := got["account_status"]; ok {
		s, _ := strconv.Atoi(fmt.Sprintf("%v", ko))
		got["account_status_f"] = accountStatuses[s]
	}

	/*
		format usd
	*/
	if ko, ok := got["balance"]; ok {
		got["balance_f"] = usd(fmt.Sprintf("%v", ko))
	}
	if ko, ok := got["amount_spent"]; ok {
		got["amount_spent_f"] = usd(fmt.Sprintf("%v", ko))
	}
	if ko, ok := got["min_daily_budget"]; ok {
		val := fmt.Sprintf("%v", ko)
		if len(val) > 3 {
			got["min_daily_budget_f"] = usd(val)
		} else {
			got["min_daily_budget_f"] = "$" + val + ".00"
		}
	}
	if ko, ok := got["spend_cap"]; ok {
		val := fmt.Sprintf("%v", ko)
		if len(val) > 3 {
			got["spend_cap_f"] = usd(val)
		} else {
			got["spend_cap_f"] = "$" + val + ".00"
		}
	}
}

func usd(val string) string {
	var v string
	for i, c := range strings.Split(val, "") {
		if i == len(val)-2 {
			v += "."
		}
		v += c
	}
	return util.StringToUsd(v)
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

	var payload struct {
		Data []interface{} `json:"data"`
		Page struct {
			Next string `json:"next"`
		} `json:"paging"`
	}

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
