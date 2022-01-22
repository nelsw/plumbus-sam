package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/model/arbo"
	"plumbus/pkg/util/logs"
	"plumbus/pkg/util/pretty"
	"time"
)

func init() {
	logs.Init()
}

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	log.WithFields(log.Fields{"ctx": ctx, "reg": req}).Info()
	switch req.RequestContext.HTTP.Method {
	case http.MethodOptions:
		return api.K()
	case http.MethodGet:
		return get(ctx)
	case http.MethodPut:
		return put()
	default:
		return api.Nada()
	}
}

func get(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodGet, arbo.URL(), nil)

	if err != nil {
		log.WithError(err).Error("new request failed")
		return api.Err(err)
	}

	for _, c := range arbo.Cookies() {
		req.AddCookie(c)
	}

	for k, v := range arbo.Headers() {
		req.Header.Set(k, v)
	}

	var res *http.Response
	if res, err = client.Do(req); err != nil {
		log.WithError(err).Error("client do failed")
		return api.Err(err)
	}

	fmt.Println(res)
	pretty.Print(res)

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			log.WithError(err).Error()
		}
	}(res.Body)

	var zr *gzip.Reader
	if zr, err = gzip.NewReader(res.Body); err != nil {
		log.WithError(err).Error("gzip reader failed")
		return api.Err(err)
	}

	var out bytes.Buffer
	var wrt int64
	if wrt, err = io.Copy(&out, zr); err != nil {
		log.WithError(err).Error("writing failed")
		return api.Err(err)
	}

	log.Trace("written ", wrt)

	fmt.Println(out.String())

	return api.K()
}

func put() (events.APIGatewayV2HTTPResponse, error) {
	return api.K()
}

func main() {
	lambda.Start(handle)
}
