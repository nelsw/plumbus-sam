package arbo

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

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
	Clicks       int         `json:"clicks"`
	Ctr          interface{} `json:"ctr"`
	Ecpc         interface{} `json:"ecpc"`
	Simpressions interface{} `json:"simpressions"`
	Revenue      int         `json:"revenue"`
	Profit       int         `json:"profit"`
	Cpm          interface{} `json:"cpm"`
	Rimpressions interface{} `json:"rimpressions"`
	Rps          int         `json:"rps"`
	Hrps         interface{} `json:"hrps"`
	Roi          string      `json:"roi"`
	Stime        string      `json:"stime"`
}

func URL() string {

	// todo - confirm in February that the date day is zero padded
	d := time.Now().Format("January 02, 2006")
	r := fmt.Sprintf("%s - %s", d, d)
	dr := url.PathEscape(r)

	// todo - confirm this can stay constant,
	// because it just so happens to be the milli epoch ts when I executed the request to write this code
	// Thursday, January 20, 2022 2:46:11.671 PM GMT-05:00
	s := "1642707971671"

	return fmt.Sprintf("https://arbotron.com/nac_handler.php?cmd=get_campaigns&dr=%s&nt=facebook&nid=all&_=%s", dr, s)
}

func Headers() map[string]string {
	return map[string]string{
		"Host":                      "arbotron.com",
		"User-Agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:96.0) Gecko/20100101 Firefox/96.0",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.5",
		"Accept-Encoding":           "gzip, deflate, br",
		"DNT":                       "1",
		"Connection":                "keep-alive",
		"Cookie":                    "AdwaveContent=ikbdhk791ak9pf66grkgqa0qq3; device_hash=07b7b9f27733b4d1cb6dae7806f20e3c127ec0a2; ajs_group_id=null; ajs_anonymous_id=%220497a316-03bc-42b7-a9fc-4af6ab3f0130%22; _hjSessionUser_2735539=eyJpZCI6ImQwNDRhOGVmLWY4NmEtNTljZC04YjkxLWYwN2M1MDEzYzM2NyIsImNyZWF0ZWQiOjE2NDE5MjA3MDQyMDYsImV4aXN0aW5nIjp0cnVlfQ==",
		"Upgrade-Insecure-Requests": "1",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
	}
}

func Cookies() []*http.Cookie {
	return []*http.Cookie{
		hjSample(),
		hjSession(),
		hjUser(),
		adwave(),
		ajsAnon(),
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

func ajsAnon() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Sat, 21 Jan 2023 23:06:04 GMT")
	return &http.Cookie{
		Name:     "ajs_anonymous_id",
		Value:    "%220497a316-03bc-42b7-a9fc-4af6ab3f0130%22",
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
