package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	log "github.com/sirupsen/logrus"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/model/account"
	"plumbus/pkg/model/arbo"
	"plumbus/pkg/model/sovrn"
	"plumbus/pkg/repo"
	"plumbus/pkg/sam"
	"plumbus/pkg/util/logs"
	"strings"
)

func init() {
	logs.Init()
}

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": req}).Info()

	data := sam.NewRequestBytes(http.MethodPut, nil)
	arboSuccess := true
	if _, err := sam.NewEvent(ctx, arbo.Handler, data); err != nil {
		log.WithError(err).Error("while invoking request response from arbo handler")
		arboSuccess = false
	}

	sovrnSuccess := true
	if err := process(ctx, req); err != nil {
		log.WithError(err).Error("while processing sovrn request")
		sovrnSuccess = false
	}

	if arboSuccess || sovrnSuccess {
		data = sam.NewRequestBytes(http.MethodPost, nil)
		if _, err := sam.NewEvent(ctx, account.Handler, data); err != nil {
			log.WithError(err).Error("while invoking account post event")
		}
	}

	// as sovrn is actively hitting this webhook,
	// we always return a 200 from this handler
	// to communicate successful delivery.
	return api.K()
}

func process(ctx context.Context, request events.APIGatewayV2HTTPRequest) (err error) {

	var pay sovrn.Payload
	if err = json.Unmarshal([]byte(strings.TrimLeft(request.Body, "attachment")), &pay); err != nil {
		log.WithError(err).Error("unable to interpret request.Body attachment payload")
		return
	}

	log.Trace("unmarshalled sovrn payload from request.Body attachment")

	var vv []sovrn.Value
	if err = json.Unmarshal([]byte(pay.Attachment.Data), &vv); err != nil {
		log.WithError(err).Error("unable to unmarshal payload attachment data into sovrn value slice")
		return err
	}

	log.WithFields(log.Fields{"size": len(vv)}).Trace("unmarshalled sovrn values from payload attachment data")

	// as each sovrn value represents a campaign,
	// we group sovrn values by (campaign) UTM
	// to sum and average value data points.
	groups := map[string][]sovrn.Value{}
	for _, v := range vv {
		if v.UTM == "" {
			log.Trace("value missing campaign id", v)
		} else if _, ok := groups[v.UTM]; ok {
			groups[v.UTM] = append(groups[v.UTM], v)
		} else {
			groups[v.UTM] = []sovrn.Value{v}
		}
	}

	log.WithFields(log.Fields{"size": len(vv)}).Trace("sovrn value groups")

	// here we perform said
	// summation and averaging,
	// before persisting to db.
	var r types.WriteRequest
	var rr []types.WriteRequest
	for utm, group := range groups {

		v := sovrn.Value{UTM: utm}

		for _, g := range group {
			v.Revenue += g.Revenue
			v.Impressions += g.Impressions
			v.Sessions += g.Sessions
			v.CTR += g.CTR
			v.PageViews += g.PageViews
		}

		v.CTR /= float64(len(group))

		if r, err = v.WriteRequest(); err != nil {
			log.WithError(err).Error("sovrn value to write request")
		} else {
			rr = append(rr, r)
		}
	}

	if err = repo.BatchWrite(ctx, sovrn.Table, rr); err != nil {
		log.WithError(err).Error("sovrn value batch write items")
		return
	}

	log.WithFields(log.Fields{"size": len(rr)}).Trace("sovrn value batch write items")

	return
}

func main() {
	lambda.Start(handle)
}
