package main

import (
	"context"
	"fmt"
	"plumbus/pkg/util"
	"testing"
)

func TestHandle(t *testing.T) {

	if out, err := handle(context.TODO()); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		fmt.Println("get accounts to ignore")
		util.PrettyPrint(out)
	}

	if out, err := handle(context.WithValue(context.TODO(), "account_id", "564715394630862")); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		fmt.Println("get account where account_id = 564715394630862")
		util.PrettyPrint(out)
	}

	if out, err := handle(context.WithValue(context.TODO(), "accounts", "")); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		fmt.Println("get all accounts")
		util.PrettyPrint(out)
	}
}
