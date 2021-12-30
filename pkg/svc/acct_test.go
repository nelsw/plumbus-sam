package svc

import (
	"plumbus/pkg/util"
	"testing"
)

func TestAccounts(t *testing.T) {
	if out, err := Accounts(); err != nil {
		t.Error(err)
	} else {
		util.PrettyPrint(out)
	}
}
