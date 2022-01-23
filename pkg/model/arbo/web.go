package arbo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func NewRequest(ctx context.Context, c Client) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, uri(c), nil)
	for _, c := range cookies() {
		req.AddCookie(c)
	}
	for k, v := range headers() {
		req.Header.Set(k, v)
	}
	return req
}

func uri(c Client) string {

	now := time.Now()

	d := now.Format("January 2, 2006")
	dr := url.PathEscape(fmt.Sprintf("%s - %s", d, d))

	s := strconv.Itoa(int(now.UnixMilli()))

	r := fmt.Sprintf("https://arbotron.com/nac_handler.php?cmd=get_campaigns&dr=%s&nt=facebook&nid=all&_=%s", dr, s)

	r = url.QueryEscape(r)

	u := fmt.Sprintf("https://arbotron.com/?cmd=switch_client&client_id=%v&redirect_url=", c)

	return u + r
}

func headers() map[string]string {

	var chunks []string
	for _, cookie := range cookies() {
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

func cookies() []*http.Cookie {
	return []*http.Cookie{
		arboHjSample(),
		arboHjSession(),
		arboHjUser(),
		arboAdwave(),
		arboAjsAnon(),
		arboAjsGroup(),
		arboDeviceHash(),
		hjZlcmid(),
		hjGa(),
		hjGid(),
		hjSessionUser(),
		hjAjsAnon(),
	}
}

func arboHjSample() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Fri, 21 Jan 2022 23:12:05 GMT")
	return &http.Cookie{
		Name:     "_hjIncludedInSessionSample",
		Value:    "0",
		Path:     "/",
		Domain:   "arbotron.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
}

func arboHjSession() *http.Cookie {
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

func arboHjUser() *http.Cookie {
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

func arboAdwave() *http.Cookie {
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

func arboAjsAnon() *http.Cookie {
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

func arboAjsGroup() *http.Cookie {
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

func arboDeviceHash() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Wed, 11 Jan 2023 17:04:59 GMT")
	return &http.Cookie{
		Name:     "device_hash",
		Value:    "07b7b9f27733b4d1cb6dae7806f20e3c127ec0a2",
		Path:     "/",
		Domain:   "arbotron.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteNoneMode,
	}
}

func hjZlcmid() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Sat, 21 Jan 2023 18:42:26 GMT")
	return &http.Cookie{
		Name:     "__zlcmid",
		Value:    "189keDwRTlMu3gh",
		Path:     "/",
		Domain:   ".hotjar.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
}

func hjGa() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Sat, 21 Jan 2024 18:42:33 GMT")
	return &http.Cookie{
		Name:     "_ga",
		Value:    "%22afdb96a9-c21d-4653-aa88-fa9bf74fa318%22",
		Path:     "/",
		Domain:   ".hotjar.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
}

func hjGid() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Sat, 22 Jan 2022 18:42:33 GMT")
	return &http.Cookie{
		Name:     "_gid",
		Value:    "%22afdb96a9-c21d-4653-aa88-fa9bf74fa318%22",
		Path:     "/",
		Domain:   ".hotjar.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
}

func hjSessionUser() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Sat, 21 Jan 2023 18:42:25 GMT")
	return &http.Cookie{
		Name:     "_hjSessionUser_605312",
		Value:    "eyJpZCI6ImE1YjEyZDVhLTQ1ZmMtNWViMy04OTQ4LWZkOTE3NmEzMmQ1NSIsImNyZWF0ZWQiOjE2NDI3OTA1NDU2NDgsImV4aXN0aW5nIjpmYWxzZX0=",
		Path:     "/",
		Domain:   ".hotjar.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
}

func hjAjsAnon() *http.Cookie {
	exp, _ := time.Parse(time.RFC1123, "Sat, 21 Jan 2023 18:42:25 GMT")
	return &http.Cookie{
		Name:     "ajs_anonymous_id",
		Value:    "%22afdb96a9-c21d-4653-aa88-fa9bf74fa318%22",
		Path:     "/",
		Domain:   ".hotjar.com",
		Expires:  exp,
		MaxAge:   int(exp.Sub(time.Now()).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
}
