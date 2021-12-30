package repo

import (
	"fmt"
	"testing"
)

func TestExists(t *testing.T) {

	if exists, err := Exists("plumbus_ignored_ad_accounts", "account_id", "564715394630862"); err != nil {
		fmt.Println(err)
		t.Fail()
	} else {
		fmt.Println(exists)
	}

}
