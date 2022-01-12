package fb

import (
	"plumbus/pkg/util"
	"testing"
)

func TestUpdateCampaignStatus(t *testing.T) {

	//id := "23850061947590705"
	//s := ActiveCampaign
	//
	//
	//if err := UpdateCampaignStatus(id, s); err != nil {
	//	t.Error(err)
	//}

	var campaignFields = "&fields=account_id,id,name,status,daily_budget,budget_remaining,created_time,updated_time"
	if data, err := Get(api + "/act_294891755970623/campaigns" + Token() + campaignFields); err != nil {
		t.Error(err)
	} else {
		util.PrettyPrint(data)
	}

}
