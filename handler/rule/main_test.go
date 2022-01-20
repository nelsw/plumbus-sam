package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"plumbus/pkg/model/rule"
	"testing"
)

func TestHandle(t *testing.T) {

	var success bool
	if success = testPut(); !success {
		t.Log("put failed")
		t.Fail()
		return
	}

	var rules []rule.Entity
	if rules = testGet(); len(rules) < 1 {
		t.Log("get failed")
		t.Fail()
		return
	}

	var out string
	if out = testDel(rules[0].ID); out != "" {
		t.Log("del failed")
		t.Log(out)
		t.Fail()
	}
}

func testPut() bool {

	b, _ := json.Marshal(&rule.Entity{
		Named: "test rule",
		Conditions: []rule.Condition{
			{
				LHS: rule.Spend,
				Op:  rule.GT,
				RHS: 100,
			},
		},
		Effect: rule.Active,
		Active: true,
	})

	req := events.APIGatewayV2HTTPRequest{
		Body: string(b),
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: http.MethodPut,
			},
		},
	}

	out, _ := handle(context.TODO(), req)

	return out.StatusCode == 200
}

func testGet() []rule.Entity {

	req := events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: http.MethodGet,
			},
		},
	}

	out, _ := handle(context.TODO(), req)

	var rules []rule.Entity

	_ = json.Unmarshal([]byte(out.Body), &rules)

	return rules
}

func testDel(id string) string {

	req := events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{"id": id},
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: http.MethodDelete,
			},
		},
	}

	out, _ := handle(context.TODO(), req)

	return out.Body
}
