package arbo

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const imgHost = "https://dwyeew221rxbg.cloudfront.net/facebook_fu/"

const (
	cbsi  = "%220497a316-03bc-42b7-a9fc-4af6ab3f0130%22"
	inuvo = "%220497a316-03bc-42b7-a9fc-4af6ab3f0130%22"
)

var (
	table   = "plumbus_arbo"
	handler = "plumbus_arboHandler"
)

func Table() string {
	return table
}

func TableName() *string {
	return &table
}

func Handler() string {
	return handler
}

type Payload struct {
	Data []Entity `json:"data"`
}

/*
 {
  "id": "365764",
  "abid": "1252743",
  "cid": "23850050568100225",
  "page_id": "112090947953846",
  "nid": "all",
  "checkbox": "365764",
  "status": "ACTIVE",
  "network": [
   "PL Content 27 - MA",
   "01-22 19:18 UTC"
  ],
  "target_url": "https://www.financerepublic.com/these-revealing-red-carpet-outfits-will-make-you-cringe-copy?utm_subid=9739674\u0026utm_adset=3672903\u0026utm_campaign=1252743\u0026utm_source=facebook\u0026utm_medium=referral",
  "img": "1a8c3d7ead46c97edff4405621fa61e8.png",
  "name": "1252743 Red Carpet - 20k - .1252743 - w - refresh - 1/16",
  "bid": "0.3100",
  "budget": "0.000",
  "buyer": "Arden",
  "spend": "2.94",
  "clicks": 23,
  "ctr": "3.470000",
  "ecpc": "0.130000",
  "simpressions": "0",
  "revenue": 2.78,
  "profit": -0.16,
  "cpm": "3.740000",
  "rimpressions": "2947",
  "rps": 0.12086956521739,
  "hrps": "0.000",
  "roi": "",
  "stime": "01-22 05:31 UTC"
 },
*/

type Entity struct {
	Id           string      `json:"id"`
	Abid         string      `json:"abid"`
	Cid          string      `json:"cid"`
	PageId       string      `json:"page_id"`
	Nid          string      `json:"nid"`
	Checkbox     string      `json:"checkbox"`
	Status       string      `json:"status"`
	Network      []string    `json:"network"`
	TargetUrl    string      `json:"target_url"`
	Img          string      `json:"img"`
	Name         string      `json:"name"`
	Bid          string      `json:"bid"`
	Budget       string      `json:"budget"`
	Buyer        string      `json:"buyer"`
	Spend        interface{} `json:"spend"`
	Clicks       interface{} `json:"clicks"`
	Ctr          interface{} `json:"ctr"`
	Ecpc         interface{} `json:"ecpc"`
	Simpressions interface{} `json:"simpressions"`
	Revenue      interface{} `json:"revenue"`
	Profit       interface{} `json:"profit"`
	Cpm          interface{} `json:"cpm"`
	Rimpressions interface{} `json:"rimpressions"`
	Rps          interface{} `json:"rps"`
	Hrps         interface{} `json:"hrps"`
	Roi          string      `json:"roi"`
	Stime        string      `json:"stime"`
}

func (e *Entity) Item() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"ID":      &types.AttributeValueMemberS{Value: e.Cid},
		"UTM":     &types.AttributeValueMemberS{Value: e.Abid},
		"Named":   &types.AttributeValueMemberS{Value: e.Name},
		"Img":     &types.AttributeValueMemberS{Value: fmt.Sprintf("%s%s", imgHost, e.Img)},
		"Revenue": &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", e.Revenue)},
	}
}

func (e *Entity) WriteRequest() types.WriteRequest {
	return types.WriteRequest{PutRequest: &types.PutRequest{Item: e.Item()}}
}

func (e *Entity) PutItemInput() *dynamodb.PutItemInput {
	return &dynamodb.PutItemInput{Item: e.Item(), TableName: TableName()}
}

func RequestCBSI(ctx context.Context) *http.Request {
	return request(ctx, cbsi)
}

func RequestInuvo(ctx context.Context) *http.Request {
	return request(ctx, inuvo)
}

func request(ctx context.Context, client string) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, uri(), nil)
	for _, c := range cookies(client) {
		req.AddCookie(c)
	}
	for k, v := range headers(client) {
		req.Header.Set(k, v)
	}
	return req
}

func uri() string {

	now := time.Now()

	d := now.Format("January 2, 2006")
	dr := url.PathEscape(fmt.Sprintf("%s - %s", d, d))

	s := strconv.Itoa(int(now.UnixMilli()))

	return fmt.Sprintf("https://arbotron.com/nac_handler.php?cmd=get_campaigns&dr=%s&nt=facebook&nid=all&_=%s", dr, s)
}

func headers(client string) map[string]string {

	var chunks []string
	for _, cookie := range cookies(client) {
		chunks = append(chunks, fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value))
	}

	return map[string]string{
		"Host":                      "arbotron.com",
		"User-Agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:96.0) Gecko/20100101 Firefox/96.0",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.5",
		"Accept-Encoding":           "gzip, deflate, br",
		"DNT":                       "1",
		"Connection":                "keep-alive",
		"Cookie":                    strings.Join(chunks, " "),
		"Upgrade-Insecure-Requests": "1",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
	}
}

func cookies(client string) []*http.Cookie {
	return []*http.Cookie{
		hjSample(),
		hjSession(),
		hjUser(),
		adwave(),
		ajsAnon(client),
		ajsGroup(),
		deviceHash(),
	}
}

func hjSample() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Fri, 21 Jan 2022 23:12:05 GMT")
	return &http.Cookie{
		Name:     "_hjIncludedInSessionSample",
		Value:    "0",
		Path:     "/",
		Domain:   ".arbotron.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
}

func hjSession() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Fri, 21 Jan 2022 23:36:04 GMT")
	return &http.Cookie{
		Name:     "_hjSession_2735539",
		Value:    "eyJpZCI6ImFjNGEyOTljLWY4OTctNDcwYi1hMzQ3LWU3ZjAyMDEyMzdmNSIsImNyZWF0ZWQiOjE2NDI4MDYzNjQzMjQsImluU2FtcGxlIjpmYWxzZX0=",
		Path:     "/",
		Domain:   ".arbotron.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
}

func hjUser() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Sat, 21 Jan 2023 23:06:04 GMT")
	return &http.Cookie{
		Name:     "_hjSessionUser_2735539",
		Value:    "eyJpZCI6ImQwNDRhOGVmLWY4NmEtNTljZC04YjkxLWYwN2M1MDEzYzM2NyIsImNyZWF0ZWQiOjE2NDE5MjA3MDQyMDYsImV4aXN0aW5nIjp0cnVlfQ==",
		Path:     "/",
		Domain:   ".arbotron.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
}

func adwave() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Sun, 11 Dec 2022 11:53:30 GMT")
	return &http.Cookie{
		Name:     "AdwaveContent",
		Value:    "ikbdhk791ak9pf66grkgqa0qq3",
		Path:     "/",
		Domain:   ".arbotron.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteNoneMode,
	}
}

func ajsAnon(client string) *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Sat, 21 Jan 2023 23:06:04 GMT")
	return &http.Cookie{
		Name:     "ajs_anonymous_id",
		Value:    client,
		Path:     "/",
		Domain:   ".arbotron.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
}

func ajsGroup() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Sat, 21 Jan 2023 23:06:04 GMT")
	return &http.Cookie{
		Name:     "ajs_group_id",
		Value:    "null",
		Path:     "/",
		Domain:   ".arbotron.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
}

func deviceHash() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Wed, 11 Jan 2023 17:04:59 GMT")
	return &http.Cookie{
		Name:     "device_hash",
		Value:    "07b7b9f27733b4d1cb6dae7806f20e3c127ec0a2",
		Path:     "/",
		Domain:   ".arbotron.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteNoneMode,
	}
}
