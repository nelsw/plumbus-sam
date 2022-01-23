package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"plumbus/pkg/model/arbo"
	"plumbus/pkg/sam"
	"plumbus/pkg/util/pretty"
	"plumbus/test"
	"testing"
)

//func TestHandlePut(t *testing.T) {
//	req := sam.NewRequest(http.MethodPut, nil)
//	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
//		t.Error(res.StatusCode, res.Body)
//	}
//}

func TestHandleGet(t *testing.T) {
	req := sam.NewRequest(http.MethodGet, nil)
	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	} else {
		var ee []arbo.Entity
		_ = json.Unmarshal([]byte(res.Body), &ee)
		pretty.Print(ee)
		fmt.Println(len(ee))
	}
}
