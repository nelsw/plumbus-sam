package fb

import (
	"fmt"
	"plumbus/pkg/util"
	"testing"
)

func TestAPI_GET(t *testing.T) {

	var out interface{}
	var err error

	fmt.Println("account insights")
	accountID := "1057633785038634"
	if out, err = Account(accountID).Insights().GET(); err != nil {
		t.Error(err)
	} else {
		util.PrettyPrint(out)
	}

	fmt.Println("campaign insights")
	if out, err = Campaigns(accountID).Insights().GET(); err != nil {
		t.Error(err)
	} else {
		util.PrettyPrint(out)
	}

	fmt.Println("adset insights")
	campaignID := "23849142462340164"
	if out, err = AdSets(campaignID).Insights().GET(); err != nil {
		t.Error(err)
	} else {
		util.PrettyPrint(out)
	}

	fmt.Println("ad insights")
	adSetID := "23849142462580164"
	if out, err = Ads(adSetID).Insights().GET(); err != nil {
		t.Error(err)
	} else {
		util.PrettyPrint(out)
	}

}
