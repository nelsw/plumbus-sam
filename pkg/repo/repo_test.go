package repo

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/ptr"
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

func TestScanInput(t *testing.T) {
	in := &dynamodb.ScanInput{TableName: ptr.String("plumbus_ignored_ad_accounts")}
	var out interface{}
	if err := ScanInputAndUnmarshal(in, &out); err != nil {
		t.Error(err)
	}
	res := map[string]interface{}{}
	for _, w := range out.([]interface{}) {
		z := w.(map[string]interface{})
		res[fmt.Sprintf("%v", z["account_id"])] = true
	}
	fmt.Println(res)
}
