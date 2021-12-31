package main

import (
	"context"
	"fmt"
	"plumbus/pkg/util"
	"testing"
)

func TestHandle(t *testing.T) {

	ctx := context.TODO()

	if out, err := handle(ctx, map[string]interface{}{"ignore": true}); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		fmt.Println("got accounts to ignore")
	}

	if out, err := handle(ctx, map[string]interface{}{"account": "564715394630862"}); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		fmt.Println("got account where account_id = 564715394630862")
	}

	if out, err := handle(ctx, map[string]interface{}{"campaigns": "1231092737389360"}); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		fmt.Println("got campaigns where account_id = 1231092737389360")
	}

	if out, err := handle(ctx, map[string]interface{}{"accounts": true}); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		util.PrettyPrint(out)
		fmt.Println("got all accounts")
	}
}
