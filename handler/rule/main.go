package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/model/campaign"
	"plumbus/pkg/model/rule"
	"plumbus/pkg/repo"
	"plumbus/pkg/sam"
	"plumbus/pkg/util/logs"
	"strings"
	"sync"
	"time"
)

func init() {
	logs.Init()
}

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": req}).Info()

	switch req.RequestContext.HTTP.Method {

	case http.MethodOptions:
		return api.K()

	case http.MethodGet:
		return get(ctx)

	case http.MethodPut:
		return put(ctx, req.Body)

	case http.MethodDelete:
		return del(ctx, req.QueryStringParameters["id"])

	case http.MethodPost:
		if _, ok := req.QueryStringParameters["all"]; ok {
			return postAll(ctx)
		} else {
			return postOne(ctx, req.Body)
		}

	default:
		return api.Nada()
	}
}

func get(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	var out []rule.Entity
	if err := repo.Scan(ctx, &dynamodb.ScanInput{TableName: rule.TableName()}, &out); err != nil {
		return api.Err(err)
	}

	return api.JSON(out)
}

func put(ctx context.Context, body string) (events.APIGatewayV2HTTPResponse, error) {

	var e rule.Entity
	if err := json.Unmarshal([]byte(body), &e); err != nil {
		return api.Err(err)
	}

	now := time.Now().UTC()
	if e.Updated = now; e.ID == "" {
		e.ID = uuid.NewString()
		e.Created = now
	}

	if item, err := attributevalue.MarshalMap(&e); err != nil {
		return api.Err(err)
	} else if err = repo.Put(ctx, &dynamodb.PutItemInput{Item: item, TableName: rule.TableName()}); err != nil {
		return api.Err(err)
	} else {
		return api.JSON(e)
	}
}

func del(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {

	in := &dynamodb.DeleteItemInput{
		TableName: rule.TableName(),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{
				Value: id,
			},
		},
	}

	if err := repo.Delete(ctx, in); err != nil {
		return api.Err(err)
	}

	return api.K()
}

func postAll(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	log.Trace("performing rule analysis requested by system")

	var ee []rule.Entity
	if err := repo.Scan(ctx, &dynamodb.ScanInput{TableName: rule.TableName()}, &ee); err != nil {
		return api.Err(err)
	}

	log.Trace("found ", len(ee), " rules to analyze and potentially act upon")

	var wg sync.WaitGroup
	for _, e := range ee {
		wg.Add(1)
		go func(e rule.Entity) {
			defer wg.Done()
			if err := post(ctx, e); err != nil {
				log.WithError(err).
					WithFields(log.Fields{"rule": e}).
					Error("while getting campaigns and/or evaluating campaigns against the given rule")
				return
			}
		}(e)
	}

	wg.Wait()

	log.Trace("completed rule analysis from user request")

	return api.K()
}

func postOne(ctx context.Context, body string) (events.APIGatewayV2HTTPResponse, error) {

	log.Trace("performing rule analysis requested by user")

	var e rule.Entity
	if err := json.Unmarshal([]byte(body), &e); err != nil {
		log.WithError(err).Error("unable to unmarshal request body into a rule entity")
		return api.Err(err)
	}

	log.WithFields(log.Fields{"rule.Entity": e}).Trace("successfully interpreted request body into a rule entity")

	if err := post(ctx, e); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"rule": e}).
			Error("while getting campaigns and/or evaluating campaigns against the given rule")
		return api.Err(err)
	}

	log.Trace("successfully completed rule analysis from user request")
	return api.K()
}

func post(ctx context.Context, r rule.Entity) error {

	var all []campaign.Entity
	for id, ids := range r.Nodes {

		params := map[string]string{"accountID": id}
		if len(ids) > 0 {
			params["campaignIDS"] = strings.Join(ids, ",")
		}

		var err error
		var out *faas.InvokeOutput
		if out, err = sam.NewReqRes(ctx, campaign.Handler(), sam.NewRequestBytes(http.MethodGet, params)); err != nil {
			return err
		}

		var res events.APIGatewayV2HTTPResponse
		if _ = json.Unmarshal(out.Payload, &res); res.StatusCode != http.StatusOK {
			return errors.New(res.Body)
		}

		var cc []campaign.Entity
		if err = json.Unmarshal([]byte(res.Body), &cc); err != nil {
			return err
		}

		all = append(all, cc...)
	}

	for _, c := range all {
		eval(ctx, r, c)
	}
	return nil
}

func eval(ctx context.Context, r rule.Entity, c campaign.Entity) {

	if string(r.Effect) == c.Circ {
		log.Trace("rule effect == campaign status")
		return
	}

	for _, condition := range r.Conditions {

		if condition.LHS == rule.Spend && condition.Met(c.Spent()) {
			log.Trace("spend met: ", c.ID)
			continue
		}

		if condition.LHS == rule.Profit && condition.Met(c.Profit) {
			log.Trace("profit met: ", c.ID)
			continue
		}

		if condition.LHS == rule.ROI && condition.Met(c.ROI) {
			log.Trace("roi met: ", c.ID)
			continue
		}

		log.Trace("nothing met: ", c.ID)
		return
	}

	log.WithFields(log.Fields{
		"AccountID":    c.AccountID,
		"CampaignID":   c.ID,
		"CampaignName": c.Name,
		"Spend":        c.Spend,
		"Revenue":      c.Revenue,
		"Profit":       c.Profit,
		"ROI":          c.ROI,
		"RuleID":       r.ID,
		"RuleName":     r.Named,
		"Rule Effect":  r.Effect,
	}).Trace("Campaign met rule conditions; Will attempt to effect change through the campaign handler")

	data := sam.NewRequestBytes(http.MethodPatch, map[string]string{
		"status":    r.Effect.String(),
		"accountID": c.AccountID,
		"ID":        c.ID,
	})

	if out, err := sam.NewEvent(ctx, campaign.Handler(), data); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"code": out.StatusCode, "payload": string(out.Payload)}).
			Error("while sending an update status event to fb handler")
		return
	}

	log.Trace("Successfully updated campaign status in Facebook and in the Plumbus database ... grab a beer.")
}

func main() {
	lambda.Start(handle)
}
