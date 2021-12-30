package main

import (
	"fmt"
	"plumbus/pkg/util"
	"testing"
)

func TestHandle(t *testing.T) {

	if out, err := handle(map[string]interface{}{"ignore": true}); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		fmt.Println("get accounts to ignore")
		util.PrettyPrint(out)
	}

	if out, err := handle(map[string]interface{}{"account": "564715394630862"}); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		fmt.Println("get account where account_id = 564715394630862")
		util.PrettyPrint(out)
	}

	if out, err := handle(map[string]interface{}{"accounts": true}); err != nil || len(out) == 0 {
		t.Fail()
	} else {
		fmt.Println("get all accounts")
		util.PrettyPrint(out)
	}
}
