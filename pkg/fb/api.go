package fb

import (
	"encoding/json"
	"fmt"
	"github.com/leekchan/accounting"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/objx"
	"io/ioutil"
	"net/http"
	"os"
	"plumbus/pkg/repo"
	"plumbus/pkg/util"
	"plumbus/pkg/util/logs"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	dateLayout     = "2006-01-02"
	timeLayout     = "15:04:05"
	datetimeLayout = dateLayout + " " + timeLayout
	zeroUSD        = "$0.00"
	zeroPCT        = "0.0%"
)

var mutex = &sync.Mutex{}
var usd = accounting.Accounting{Symbol: "$", Precision: 2}
var dec = accounting.Accounting{Symbol: "", Precision: 2}
var num = accounting.Accounting{}
var pct = accounting.Accounting{Symbol: "%", Precision: 2, Format: "%v%s"}

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
	accountInsightsFields = []string{
		"account_id",
		"date_start",
		"date_stop",
		"impressions",
		"clicks",
		"spend",
	}
	campaignMarketingFields = []string{
		"id",
		"account_id",
		"budget_remaining",
		"created_time",
		"name",
		"start_time",
		"status",
		"stop_time",
		"updated_time",
	}
	campaignInsightsFields = []string{
		"account_id",
		"campaign_id",
		"campaign_name",
		"impressions",
		"clicks",
		"spend",
		"cpc",
		"cpm",
		"cpp",
		"ctr",
		"created_time",
		"updated_time",
	}
	adSetInsightsFields = []string{
		"account_id",
		"campaign_id",
		"adset_id",
		"adset_name",
		"impressions",
		"clicks",
		"spend",
		"cpc",
		"cpm",
		"cpp",
		"ctr",
		"created_time",
		"updated_time",
	}
	adInsightsFields = []string{
		"account_id",
		"campaign_id",
		"adset_id",
		"ad_id",
		"ad_name",
		"impressions",
		"clicks",
		"spend",
		"cpc",
		"cpm",
		"cpp",
		"ctr",
		"created_time",
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
	adSets,
	ads,
	marketing,
	insights bool

	dateYesterday,
	dateToday,
	dateTomorrow string

	accountID,
	campaignID,
	adSetID,
	adID,
	url string

	fields []string

	ignore map[string]interface{}

	initDatetime,
	startDatetime time.Time
	startHour int

	startYear, startDay int
	startMonth          time.Month
}

func init() {
	logs.Init()
}

func newAPI() *API {
	a := new(API)
	a.initDatetime = time.Now().Local()
	a.startDatetime = a.initDatetime.Add(time.Hour * 24 * -1).Truncate(time.Hour)
	a.startHour, _ = strconv.Atoi(a.startDatetime.Format("15"))
	y, m, d := a.startDatetime.Date()
	a.startYear = y
	a.startMonth = m
	a.startDay = d
	return a
}

func Accounts(ignore map[string]interface{}) *API {
	a := newAPI()
	a.account = true
	a.ignore = ignore
	return a
}

func Account(id string) *API {
	a := newAPI()
	a.account = true
	a.accountID = id
	return a
}

func Campaign(id string) *API {
	a := newAPI()
	a.campaignID = id
	a.campaign = true
	return a
}

func Campaigns(id string) *API {
	a := newAPI()
	a.accountID = id
	a.campaign = true
	return a
}

func AdSets(id string) *API {
	a := newAPI()
	a.campaignID = id
	a.adSets = true
	return a
}

func Ads(id string) *API {
	a := newAPI()
	a.adSetID = id
	a.ads = true
	return a
}

func (a *API) Fields(fields []string) *API {
	a.fields = append(a.fields, fields...)
	return a
}

func (a *API) Insights() *API {
	a.insights = true
	if a.account {
		return a.Fields(accountInsightsFields)
	} else if a.campaign {
		return a.Fields(campaignInsightsFields)
	} else if a.adSets {
		return a.Fields(adSetInsightsFields)
	} else if a.ads {
		return a.Fields(adInsightsFields)
	} else {
		return a
	}
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

type CampaignPayload struct {
	AccountID    string `json:"account_id"`
	CampaignID   string `json:"campaign_id"`
	CampaignUTM  string `json:"campaign_utm"`
	CampaignName string `json:"campaign_name"`

	Spend  float64 `json:"spend"`
	SpendF string  `json:"spend_f"`

	Revenue  float64 `json:"revenue"`
	RevenueF string  `json:"revenue_f"`

	Profit  float64 `json:"profit"`
	ProfitF string  `json:"profit_f"`

	SovrnCTR  float64 `json:"sovrn_ctr"`
	SovrnCTRF string  `json:"sovrn_ctr_f"`

	FBCTR  float64 `json:"fb_ctr"`
	FBCTRF string  `json:"fb_ctr_f"`

	CPC  float64 `json:"cpc"`
	CPCF string  `json:"cpc_f"`

	CPM  float64 `json:"cpm"`
	CPMF string  `json:"cpm_f"`

	CPP  float64 `json:"cpp"`
	CPPF string  `json:"cpp_f"`

	Sessions  float64 `json:"sessions"`
	SessionsF string  `json:"sessions_f"`

	PageViews  float64 `json:"page_views"`
	PageViewsF string  `json:"page_views_f"`

	FBImpressions  int    `json:"fb_impressions"`
	FBImpressionsF string `json:"fb_impressions_f"`

	SovrnImpressions  float64 `json:"sovrn_impressions"`
	SovrnImpressionsF string  `json:"sovrn_impressions_f"`

	Created string `json:"created"`
	Updated string `json:"updated"`

	RevenueTracked bool `json:"revenue_tracked"`

	Data []CampaignData `json:"data"`
}

func NewCampaignPayload(data []CampaignData) CampaignPayload {

	sort.Sort(ByDatetime(data))

	z := data[0]

	p := CampaignPayload{
		AccountID:    z.AccountID,
		CampaignID:   z.CampaignID,
		CampaignName: z.CampaignName,
		Created:      z.Created,
		Updated:      z.Updated,
		Data:         data,
	}

	for _, d := range data {
		p.Spend += d.Spend
		p.FBImpressions += d.Impressions
		p.FBCTR += d.CTR
		p.CPC += d.CPC
		p.CPM += d.CPM
		p.CPP += d.CPP
	}

	if p.SpendF = zeroUSD; p.Spend > 0 {
		p.SpendF = usd.FormatMoney(p.Spend)
	}

	if p.FBImpressionsF = "0"; p.FBImpressions > 0 {
		p.FBImpressionsF = dec.FormatMoney(p.FBImpressions)
	}

	qty := decimal.NewFromFloat(float64(len(data)))

	if p.FBCTRF = zeroPCT; p.FBCTR > 0 {
		p.FBCTRF = pct.FormatMoneyDecimal(decimal.NewFromFloat(p.FBCTR).Div(qty).Truncate(2))
	}

	if p.CPMF = zeroUSD; p.CPM > 0 {
		p.CPMF = usd.FormatMoney(decimal.NewFromFloat(p.CPM).Div(qty).Truncate(2))
	}

	if p.CPCF = zeroUSD; p.CPC > 0 {
		p.CPCF = usd.FormatMoney(decimal.NewFromFloat(p.CPC).Div(qty).Truncate(2))
	}

	if p.CPPF = zeroUSD; p.CPP > 0 {
		p.CPPF = usd.FormatMoney(decimal.NewFromFloat(p.CPP).Div(qty).Truncate(2))
	}

	if subbed := getParenWrappedCampaignUTM(p.CampaignName); util.IsNumber(subbed) {
		p.CampaignUTM = subbed
	} else if spaced := strings.Split(p.CampaignName, " "); len(spaced) > 1 && util.IsNumber(spaced[0]) {
		p.CampaignUTM = spaced[0]
	} else if scored := strings.Split(p.CampaignName, "_"); len(scored) > 1 && util.IsNumber(scored[0]) {
		p.CampaignUTM = scored[0]
	} else {
		p.CampaignUTM = p.CampaignID
	}

	var err error
	var bytes []byte

	if bytes, err = repo.Get("plumbus_fb_sovrn", "campaign", p.CampaignUTM); err != nil {
		log.Trace(err)
		return p
	}

	var m map[string]interface{}
	if err = json.Unmarshal(bytes, &m); err != nil {
		log.WithError(err).Error()
		return p
	}

	if p.RevenueTracked = len(m) > 0; !p.RevenueTracked {
		log.Trace("missing revenue for campaign ", p)
		return p
	}

	if p.RevenueF = zeroUSD; m["revenue"] != nil {
		if p.Revenue = m["revenue"].(float64); p.Revenue > 0 {
			p.RevenueF = usd.FormatMoneyFloat64(p.Revenue)
		}
	}

	if p.SovrnCTRF = zeroPCT; m["ctr"] != nil {
		if p.SovrnCTR = m["ctr"].(float64); p.SovrnCTR > 0 {
			pct.FormatMoneyFloat64(p.SovrnCTR)
		}
	}

	if p.SessionsF = zeroUSD; m["sessions"] != nil {
		if p.Sessions = m["sessions"].(float64); p.Sessions > 0 {
			p.SessionsF = num.FormatMoney(p.Sessions)
		}
	}

	if p.PageViewsF = zeroUSD; m["page_views"] != nil {
		if p.PageViews = m["page_views"].(float64); p.PageViews > 0 {
			p.PageViewsF = num.FormatMoney(p.PageViews)
		}
	}

	if p.SovrnImpressionsF = "0"; m["impressions"] != nil {
		if p.SovrnImpressions = m["impressions"].(float64); p.SovrnImpressions > 0 {
			p.SovrnImpressionsF = num.FormatMoneyFloat64(p.SovrnImpressions)
		}
	}

	if p.Profit = p.Revenue - p.Spend; p.Profit != 0 {
		p.ProfitF = usd.FormatMoney(p.Profit)
	} else {
		p.ProfitF = zeroUSD
	}

	return p
}

type CampaignData struct {
	AccountID    string    `json:"-"`
	CampaignID   string    `json:"-"`
	CampaignName string    `json:"-"`
	Clicks       int       `json:"clicks"`
	ClicksF      string    `json:"clicks_f"`
	CTR          float64   `json:"ctr"`
	CTRF         string    `json:"ctr_f"`
	CPC          float64   `json:"cpc"`
	CPCF         string    `json:"cpc_f"`
	CPM          float64   `json:"cpm"`
	CPMF         string    `json:"cpm_f"`
	CPP          float64   `json:"cpp"`
	CPPF         string    `json:"cpp_f"`
	Impressions  int       `json:"impressions"`
	ImpressionsF string    `json:"impressions_f"`
	Spend        float64   `json:"spend"`
	SpendF       string    `json:"spend_f"`
	Datetime     time.Time `json:"datetime"`
	Created      string    `json:"-"`
	Updated      string    `json:"-"`
}

func NewCampaignData(o interface{}) CampaignData {

	m := o.(map[string]interface{})

	d := CampaignData{
		AccountID:    m["account_id"].(string),
		CampaignID:   m["campaign_id"].(string),
		CampaignName: m["campaign_name"].(string),
		Created:      m["created_time"].(string),
		Updated:      m["updated_time"].(string),
	}

	var err error

	dateStart := m["date_start"].(string)
	span := m["hourly_stats_aggregated_by_advertiser_time_zone"].(string)
	timeStart := strings.Split(span, " - ")[0]
	if d.Datetime, err = time.ParseInLocation(datetimeLayout, dateStart+" "+timeStart, time.Local); err != nil {
		log.Trace(err)
	}

	if d.CTRF = zeroPCT; m["ctr"] != nil {
		if d.CTR, err = strconv.ParseFloat(m["ctr"].(string), 64); err != nil {
			log.Trace(err)
		}
		if d.CTR > 0 {
			d.CTRF = pct.FormatMoneyFloat64(d.CTR)
		}
	}

	if d.CPCF = zeroUSD; m["cpc"] != nil {
		if d.CPC, err = strconv.ParseFloat(m["cpc"].(string), 64); err != nil {
			log.Trace(err)
		}
		if d.CPC > 0 {
			d.CPCF = usd.FormatMoneyFloat64(d.CPC)
		}
	}

	if d.ClicksF = "0"; m["clicks"] != nil {
		if d.Clicks, err = strconv.Atoi(m["clicks"].(string)); err != nil {
			log.Trace(err)
		}
		if d.Clicks > 0 {
			d.ClicksF = num.FormatMoneyInt(d.Clicks)
		}
	}

	if d.CPMF = zeroUSD; m["cpm"] != nil {
		if d.CPM, err = strconv.ParseFloat(m["cpm"].(string), 64); err != nil {
			log.Trace(err)
		}
		if d.CPM > 0 {
			d.CPMF = usd.FormatMoneyFloat64(d.CPM)
		}
	}

	if d.CPPF = zeroUSD; m["cpp"] != nil {
		if d.CPP, err = strconv.ParseFloat(m["cpp"].(string), 64); err != nil {
			log.Trace(err)
		}
		if d.CPP > 0 {
			d.CPPF = usd.FormatMoneyFloat64(d.CPP)
		}
	}

	if d.SpendF = zeroUSD; m["spend"] != nil {
		if d.Spend, err = strconv.ParseFloat(m["spend"].(string), 64); err != nil {
			log.Trace(err)
		}
		if d.Spend > 0 {
			d.SpendF = usd.FormatMoneyFloat64(d.Spend)
		}
	}

	if d.ImpressionsF = "0"; m["impressions"] != nil {
		if d.Impressions, err = strconv.Atoi(m["impressions"].(string)); err != nil {
			log.Trace(err)
		}
		if d.Impressions > 0 {
			d.ImpressionsF = num.FormatMoneyInt(d.Impressions)
		}
	}

	return d
}

type ByDatetime []CampaignData

func (a ByDatetime) Len() int           { return len(a) }
func (a ByDatetime) Less(i, j int) bool { return a[i].Datetime.Before(a[j].Datetime) }
func (a ByDatetime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (a *API) GET() (map[string]interface{}, error) {

	a.setURL()

	if a.insights {

		if a.ads {
			a.set(a.adSetID)
		} else if a.adSets {
			a.set(a.campaignID)
		} else {
			a.setAdAccount()
		}
		a.
			setInsights().
			setToken().
			setInsightLevel().
			setBreakdowns().
			setFields().
			setTimeRanges()

		out, err := get(a.url)
		if err != nil {
			log.WithError(err)
			return nil, err
		}

		var wg1 sync.WaitGroup

		campaignDataMap := map[string][]CampaignData{}
		for _, o := range out {

			wg1.Add(1)

			go func(o interface{}) {

				defer wg1.Done()

				data := NewCampaignData(o)

				if data.Datetime.Before(a.startDatetime) {
					return
				}

				mutex.Lock()
				if _, ok := campaignDataMap[data.CampaignID]; !ok {
					campaignDataMap[data.CampaignID] = []CampaignData{data}
				} else {
					campaignDataMap[data.CampaignID] = append(campaignDataMap[data.CampaignID], data)
				}
				mutex.Unlock()
			}(o)
		}

		wg1.Wait()

		if len(campaignDataMap) == 0 {
			return nil, nil
		}

		var wg2 sync.WaitGroup

		got := map[string]interface{}{
			"account_revenue": float64(0),
			"account_spend":   float64(0),
			"account_profit":  float64(0),
		}
		var campaignPayloads []CampaignPayload
		for campaignID, campaigns := range campaignDataMap {
			wg2.Add(1)
			go func(campaignID string, campaigns []CampaignData) {
				defer wg2.Done()
				p := NewCampaignPayload(campaigns)
				campaignPayloads = append(campaignPayloads, p)
				mutex.Lock()
				got["account_revenue"] = got["account_revenue"].(float64) + p.Revenue
				got["account_spend"] = got["account_spend"].(float64) + p.Spend
				got["account_profit"] = got["account_profit"].(float64) + p.Profit
				mutex.Unlock()
			}(campaignID, campaigns)
		}

		wg2.Wait()

		got["campaigns"] = campaignPayloads

		if got["account_revenue_f"] = zeroUSD; got["account_revenue"] != nil && got["account_revenue"].(float64) > 0 {
			got["account_revenue_f"] = usd.FormatMoney(got["account_revenue"])
		}

		if got["account_spend_f"] = zeroUSD; got["account_spend"] != nil && got["account_spend"].(float64) > 0 {
			got["account_spend_f"] = usd.FormatMoney(got["account_spend"])
		}

		if got["account_profit_f"] = zeroUSD; got["account_profit"] != nil && got["account_profit"].(float64) > 0 {
			got["account_profit_f"] = usd.FormatMoney(got["account_profit"])
		}

		return got, nil
	}

	if a.account && a.accountID == "" {
		return a.getAdAccounts()
	}

	if a.campaign && a.accountID != "" {

		a.setAdAccount()
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

			id := objx.New(o).Get("id").String()
			m := objx.New(o).Value().ObjxMap()
			name := fmt.Sprintf("%v", m["name"])
			remainingBudget := fmt.Sprintf("%v", m["budget_remaining"])

			if len(remainingBudget) > 3 {
				m["budget_remaining_f"] = usd.FormatMoney(remainingBudget)
			} else {
				m["budget_remaining_f"] = "$" + remainingBudget + ".00"
			}

			var campaignUTM string
			if subbed := getParenWrappedCampaignUTM(name); util.IsNumber(subbed) {
				campaignUTM = subbed
			} else if spaced := strings.Split(name, " "); len(spaced) > 1 && util.IsNumber(spaced[0]) {
				campaignUTM = spaced[0]
			} else if scored := strings.Split(name, "_"); len(scored) > 1 && util.IsNumber(scored[0]) {
				campaignUTM = scored[0]
			} else {
				campaignUTM = id
			}
			m["utm_campaign"] = campaignUTM

			var bytes []byte
			if bytes, err = repo.Get("plumbus_fb_sovrn", "campaign", campaignUTM); err != nil {
				log.WithError(err).Error()
			} else {
				var payload map[string]interface{}
				if err = json.Unmarshal(bytes, &payload); err != nil {
					log.WithError(err).Error()
				} else {
					m["revenue"] = payload["revenue"]
					m["impressions"] = payload["impressions"]
					m["sessions"] = payload["sessions"]
					m["ctr"] = payload["ctr"]
					m["page_views"] = payload["page_views"]
				}
			}

			got[id] = m
		}

		return got, nil
	}

	if a.accountID != "" {
		a.setAdAccount()
	} else if a.campaignID != "" {
		a.url += "/" + a.campaignID
	} else if a.adSetID != "" {
		a.url += "/" + a.adSetID
	} else {
		a.url += "/" + a.adID
	}

	a.setToken()
	a.setFields()

	var err error
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

func (a *API) setURL() *API {
	a.url = "https://graph.facebook.com/v12.0"
	return a
}

func (a *API) setUserID() *API {
	a.url += "/10158615602243295"
	return a
}

func (a *API) setInsights() *API {
	a.url += "/insights"
	return a
}

func (a *API) setAdAccounts() {
	a.url += "/adaccounts"
}

func (a *API) setAdAccount() *API {
	a.url += "/act_" + a.accountID
	return a
}

func (a *API) set(s string) *API {
	a.url += "/" + s
	return a
}

func (a *API) setCampaigns() {
	a.url += "/campaigns"
}

func (a *API) setToken() *API {
	a.url += "?access_token=" + os.Getenv("tkn")
	return a
}

func (a *API) setBreakdowns() *API {
	a.url += "&breakdowns=hourly_stats_aggregated_by_advertiser_time_zone"
	return a
}

func (a *API) setTimeRanges() *API {
	n := time.Now().Local()
	a.dateYesterday = n.Add(time.Hour * 24 * -1).Format(dateLayout)
	a.dateToday = n.Format(dateLayout)
	a.dateTomorrow = n.Add(time.Hour * 24).Format(dateLayout)
	b1 := "{since:'" + a.dateYesterday + "',until:'" + a.dateToday + "'}"
	b2 := "{since:'" + a.dateToday + "',until:'" + a.dateTomorrow + "'}"
	a.url += "&time_ranges=[" + strings.Join([]string{b1, b2}, ",") + "]"
	return a
}

func (a *API) setInsightLevel() *API {
	var level string
	if a.ads {
		level = "ad"
	} else if a.adSets {
		level = "adset"
	} else if a.campaign {
		level = "campaign"
	} else {
		level = "account"
	}
	a.url += "&level=" + level
	return a
}

func (a *API) setFields() *API {
	if len(a.fields) > 0 {
		a.url += "&fields=" + strings.Join(a.fields, ",")
	}
	return a
}

func (a *API) getAdAccounts() (map[string]interface{}, error) {

	a.setUserID()
	a.setAdAccounts()
	a.setToken()
	a.setFields()

	var err error
	var out []interface{}
	if out, err = get(a.url); err != nil {
		log.WithError(err).Error()
		return nil, err
	}

	res := map[string]interface{}{}
	for _, w := range out {
		z := w.(map[string]interface{})
		accountID := fmt.Sprintf("%v", z["account_id"])
		if _, ok := a.ignore[accountID]; !ok {
			res[accountID] = z
		}

	}
	return res, nil

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
		got["balance_f"] = fbUSD(ko)
	}
	if ko, ok := got["amount_spent"]; ok {
		got["amount_spent_f"] = fbUSD(ko)
	}
	if ko, ok := got["min_daily_budget"]; ok {
		val := fmt.Sprintf("%v", ko)
		if val == "100" {
			got["min_daily_budget_f"] = "$1.00"
		} else if len(val) > 3 {
			got["min_daily_budget_f"] = fbUSD(val)
		} else {
			got["min_daily_budget_f"] = "$" + val + ".00"
		}
	}
	if ko, ok := got["spend_cap"]; ok {
		val := fmt.Sprintf("%v", ko)
		if len(val) > 3 {
			got["spend_cap_f"] = fbUSD(val)
		} else {
			got["spend_cap_f"] = "$" + val + ".00"
		}
	}
}

type Payload struct {
	Data []interface{} `json:"data"`
	Page struct {
		Next string `json:"next"`
	} `json:"paging"`
}

func fbUSD(i interface{}) string {
	val := i.(string)
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

	var payload Payload
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

func getParenWrappedCampaignUTM(s string) string {
	if chunks := strings.Split(s, "("); len(chunks) > 1 {
		if chunks = strings.Split(chunks[1], ")"); len(chunks) > 1 {
			return chunks[0]
		}
	}
	return ""
}
