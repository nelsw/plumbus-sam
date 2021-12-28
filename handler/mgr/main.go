package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	mutex      = &sync.Mutex{}
	digitCheck = regexp.MustCompile(`^[0-9]+$`)
)

func main() {
	getAllAdAccounts()
}

type AdAccount struct {
	ID          string  `json:"id"`
	AccountID   string  `json:"account_id"`
	Name        string  `json:"name"`
	Age         float64 `json:"age"`
	AmountSpent string  `json:"amount_spent"`
	Status      int     `json:"account_status"`
}

type Campaign struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
}

type AdSet struct {
	CampaignID   string `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`
	AccountID    string `json:"account_id"`
	AccountName  string `json:"account_name"`
	Spend        string `json:"spend"`
}

type CampaignGuard struct {
	FacebookCampaignID string  `json:"facebook_campaign_id"`
	SovrnCampaignID    string  `json:"sovrn_campaign_id"`
	Spend              float64 `json:"spend"`
}

func getAllAdAccounts() {

	base := "https://graph.facebook.com/v12.0/10158615602243295/adaccounts"
	token := "?access_token=" + os.Getenv("tkn")
	fields := "&fields=account_id,name,age,amount_spent,account_status"
	url := base + token + fields

	var err error
	var data []AdAccount

	if data, err = getAdAccounts(url); err != nil {
		panic(err)
	}

	var activeAdAccounts []AdAccount
	for _, adAccount := range data {
		if (adAccount.Status == 1 || adAccount.Status == 201) && adAccount.AmountSpent != "0" {
			activeAdAccounts = append(activeAdAccounts, adAccount)
		}
	}

	actives := getActiveCampaigns(activeAdAccounts)

	var f float64
	var guards []CampaignGuard
	for accountID, _ := range actives {

		for _, adSet := range getAllAdSetInsights(accountID) {

			if f, err = strconv.ParseFloat(adSet.Spend, 64); err != nil || f < 100 {
				if err != nil {
					PrettyPrint(adSet)
					fmt.Println(err)
				}
				continue
			}

			sovrnCampaignID := adSet.CampaignID
			if spaced := strings.Split(adSet.CampaignName, " "); len(spaced) > 1 && digitCheck.MatchString(spaced[0]) {
				sovrnCampaignID = spaced[0]
			} else if scored := strings.Split(adSet.CampaignName, "_"); len(scored) > 1 && digitCheck.MatchString(scored[0]) {
				sovrnCampaignID = scored[0]
			}

			guards = append(guards, CampaignGuard{adSet.CampaignID, sovrnCampaignID, f})
		}
	}

	sort.SliceStable(guards, func(i, j int) bool {
		return guards[i].Spend > guards[j].Spend
	})

	PrettyPrint(guards)
	PrettyPrint(len(guards))
}

func getActiveCampaigns(activeAdAccounts []AdAccount) map[string][]Campaign {
	var wg sync.WaitGroup

	actives := map[string][]Campaign{}
	for _, activeAdAccount := range activeAdAccounts {
		wg.Add(1)
		go func(activeAdAccount AdAccount) {
			defer wg.Done()
			activeCampaigns := getAllCampaignsUnderAccount(activeAdAccount.AccountID)
			if len(activeCampaigns) > 0 {
				mutex.Lock()
				actives[activeAdAccount.AccountID] = activeCampaigns
				mutex.Unlock()
			}
		}(activeAdAccount)
	}

	wg.Wait()

	return actives
}

func getAdAccounts(url string) (data []AdAccount, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}

	var payload struct {
		Data []AdAccount `json:"data"`
		Page struct {
			Next string `json:"next"`
		} `json:"paging"`
	}

	if err = json.Unmarshal(body, &payload); err != nil {
		return
	}

	if data = append(payload.Data); payload.Page.Next == "" {
		return
	}

	var next []AdAccount
	if next, err = getAdAccounts(payload.Page.Next); err != nil {
		return
	}

	data = append(data, next...)
	return data, nil
}

func getAllCampaignsUnderAccount(account string) []Campaign {

	base := "https://graph.facebook.com/v12.0/act_" + account + "/campaigns"
	token := "?access_token=" + os.Getenv("tkn")
	fields := "&fields=id,account_id,name,status"
	url := base + token + fields

	var err error
	var data []Campaign

	if data, err = getCampaigns(url); err != nil {
		panic(err)
	}

	var activeCampaigns []Campaign
	for _, d := range data {
		if d.Status == "ACTIVE" {
			activeCampaigns = append(activeCampaigns, d)
		}
	}
	return activeCampaigns
}

func getCampaigns(url string) (data []Campaign, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}

	var payload struct {
		Data []Campaign `json:"data"`
		Page struct {
			Next string `json:"next"`
		} `json:"paging"`
	}

	if err = json.Unmarshal(body, &payload); err != nil {
		return
	}

	if data = append(payload.Data); payload.Page.Next == "" {
		return
	}

	var next []Campaign
	if next, err = getCampaigns(payload.Page.Next); err != nil {
		return
	}

	data = append(data, next...)
	return data, nil
}

func getAllAdSetInsights(adSet string) []AdSet {

	base := "https://graph.facebook.com/v12.0/act_" + adSet + "/insights"
	token := "?access_token=" + os.Getenv("tkn")
	fields := "&fields=ad_id,ad_name,adset_id,adset_name,campaign_id,campaign_name,account_id,account_name,spend"
	date := "&date_preset=today"
	level := "&level=campaign"
	url := base + token + fields + date + level

	var err error
	var data []AdSet

	if data, err = getAdSetInsights(url); err != nil {
		panic(err)
	}

	return data
}

func getAdSetInsights(url string) (data []AdSet, err error) {

	var res *http.Response
	if res, err = http.Get(url); err != nil {
		return
	}

	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}

	var payload struct {
		Data []AdSet `json:"data"`
		Page struct {
			Next string `json:"next"`
		} `json:"paging"`
	}

	if err = json.Unmarshal(body, &payload); err != nil {
		return
	}

	if data = append(payload.Data); payload.Page.Next == "" {
		return
	}

	var next []AdSet
	if next, err = getAdSetInsights(payload.Page.Next); err != nil {
		return
	}

	data = append(data, next...)
	return data, nil
}

func Pretty(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "    ")
	return string(b)
}

func PrettyPrint(v interface{}) {
	fmt.Println(Pretty(v))
}
