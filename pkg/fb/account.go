package fb

import (
	log "github.com/sirupsen/logrus"
	"plumbus/pkg/util"
)

type Identity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AccountRoot struct {
	Accounts []AccountNode `json:"accounts"`
}

type AccountNode struct {
	Identity
	Campaigns []CampaignNode `json:"children"`
}

type CampaignNode struct {
	Identity
}

func AccountTree(ignore map[string]interface{}) (got map[string]interface{}, err error) {

	url1 := api + "/" + getUserID() + "/" + getAdAccounts() + getTokenPart()

	var out []interface{}
	if out, err = get(url1); err != nil {
		log.WithError(err).Error()
		return
	}

	util.PrettyPrint(out)

	return
}
