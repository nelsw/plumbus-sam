package main

import (
	"plumbus/pkg/util"
	"testing"
)

func TestHandle(t *testing.T) {
	if accounts, err := handle(); err != nil || len(accounts) == 0 {
		t.Fail()
	} else {
		util.PrettyPrint(accounts)
	}
}
