// Package provides functionality for updating and return account entity data.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	faas "github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go/ptr"
	log "github.com/sirupsen/logrus"
	"net/http"
	"plumbus/pkg/api"
	"plumbus/pkg/model/account"
	"plumbus/pkg/model/campaign"
	"plumbus/pkg/model/fb"
	"plumbus/pkg/repo"
	"plumbus/pkg/sam"
	"plumbus/pkg/util/logs"
	"regexp"
	"sort"
	"sync"
)

var posRegexp = regexp.MustCompile(`all|in`)

func init() {
	logs.Init()
}

func handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	log.WithFields(log.Fields{"ctx": ctx, "req": req}).Info()

	switch req.RequestContext.HTTP.Method {

	case http.MethodOptions:
		return api.K()

	case http.MethodGet:
		return get(ctx, req.QueryStringParameters["pos"])

	case http.MethodPut:
		return put(ctx)

	case http.MethodPatch:
		return patch(ctx, req.QueryStringParameters["id"])

	default:
		return api.Nada()
	}

}

// get scans the db for all accounts where the included value is true, false, or either.
func get(ctx context.Context, pos string) (events.APIGatewayV2HTTPResponse, error) {

	if !posRegexp.MatchString(pos) {
		return api.Err(errors.New("unknown pos: " + pos))
	}

	in := dynamodb.ScanInput{TableName: account.TableName()}
	if pos != "all" {
		in.FilterExpression = ptr.String("Included = :v1")
		in.ExpressionAttributeValues = map[string]types.AttributeValue{
			":v1": &types.AttributeValueMemberBOOL{Value: true},
		}
	}

	var aa []account.Entity
	if err := repo.Scan(ctx, &in, &aa); err != nil {
		return api.Err(err)
	}

	sort.Sort(account.ByName(aa))

	if pos == "all" {
		if bytes, err := json.Marshal(&aa); err != nil {
			return api.Err(err)
		} else {
			return api.OK(string(bytes))
		}
	}

	var wg sync.WaitGroup

	for i, a := range aa {

		if !a.Included {
			continue
		}

		wg.Add(1)

		go func(i int, a account.Entity) {

			defer wg.Done()

			var err error
			var out *faas.InvokeOutput
			var data = sam.NewRequestBytes(http.MethodGet, map[string]string{"accountID": a.ID})
			if out, err = sam.NewReqRes(ctx, campaign.Handler(), data); err != nil {
				log.WithError(err).Error()
				return
			}

			var res events.APIGatewayV2HTTPResponse
			if err = json.Unmarshal(out.Payload, &res); err != nil {
				log.WithError(err).Error()
				return
			}

			if res.StatusCode == http.StatusNotFound {
				return
			}

			if res.StatusCode != http.StatusOK {
				log.Warn("unable to get campaigns for account ", a.ID, " status code ", res.StatusCode)
				return
			}

			var cc []campaign.Entity
			if err = json.Unmarshal([]byte(res.Body), &cc); err != nil {
				log.WithError(err).Error()
				return
			}

			aa[i].Children = cc
		}(i, a)
	}

	wg.Wait()

	if bytes, err := json.Marshal(&aa); err != nil {
		return api.Err(err)
	} else {
		return api.OK(string(bytes))
	}
}

// put requests all accounts from the FB handler and reconciles them with the db before returning given results.
func put(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {

	var err error
	var out *faas.InvokeOutput

	data, _ := json.Marshal(map[string]string{"node": "accounts"})
	if out, err = sam.NewReqRes(ctx, fb.Handler(), data); err != nil {
		return api.Err(err)
	}

	var aa []account.Entity
	if err = json.Unmarshal(out.Payload, &aa); err != nil {
		return api.Err(err)
	}

	var x account.Entity
	for _, a := range aa {

		if err = repo.Get(ctx, account.Table(), "ID", a.ID, &x); err != nil {
			return api.Err(err)
		}

		if x.ID != "" && a.Named == x.Named {
			continue
		}

		if err = repo.Put(ctx, x.PutItemInput()); err != nil {
			return api.Err(err)
		}
	}

	return get(ctx, "all")
}

// patch will toggle account inclusion.
// As an HTTP Request method, Patch is like Put without guaranteeing idempotence.
// Read more here: https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/PATCH
func patch(ctx context.Context, id string) (events.APIGatewayV2HTTPResponse, error) {

	var x account.Entity
	if err := repo.Get(ctx, account.Table(), "ID", id, &x); err != nil {
		return api.Err(err)
	}

	x.Included = !x.Included
	if err := repo.Put(ctx, x.PutItemInput()); err != nil {
		return api.Err(err)
	}

	return api.K()
}

func main() {
	lambda.Start(handle)
}
