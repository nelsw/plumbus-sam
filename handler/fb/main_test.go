package main

import (
	"context"
	"plumbus/pkg/util/pretty"
	"testing"
)

var ctx = context.TODO()

func TestHandle(t *testing.T) {
	if res, err := handle(ctx, map[string]interface{}{"node": "account"}); err != nil {
		t.Error(err)
	} else {
		pretty.PrintJson(res)
	}
}
