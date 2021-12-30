package main

import (
	"context"
	"plumbus/pkg/util"
	"testing"
)

func TestHandle(t *testing.T) {

	ctx := context.WithValue(context.TODO(), "account_id", "564715394630862")
	if out, err := handle(ctx); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		util.PrettyPrint(out)
	}

	ctx = context.WithValue(context.TODO(), "campaign_id", "23849423300680072")
	if out, err := handle(ctx); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		util.PrettyPrint(out)
	}
}
