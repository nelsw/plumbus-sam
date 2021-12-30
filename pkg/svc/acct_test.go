package svc

import (
	"plumbus/pkg/util"
	"testing"
)

func TestAccounts(t *testing.T) {
	if out, err := Accounts(); err != nil || len(out) == 0 {
		t.Error(err)
	} else {
		util.PrettyPrint(out)
	}
}
