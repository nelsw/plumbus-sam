package main

import (
	"plumbus/pkg/util"
	"testing"
)

func TestHandle(t *testing.T) {
	if accounts, err := handle(); err != nil || accounts == nil {
		t.Fail()
	} else {
		util.PrettyPrint(accounts)
	}
}
