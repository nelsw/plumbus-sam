package main

import (
	"net/http"
	"plumbus/pkg/sam"
	"plumbus/test"
	"testing"
)

//func TestHandleGetAll(t *testing.T) {
//	req := sam.NewRequest(http.MethodGet, map[string]string{"pos": "all"})
//	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
//		t.Error(res.StatusCode, res.Body)
//	}
//}
//
//func TestHandleGetIn(t *testing.T) {
//	req := sam.NewRequest(http.MethodGet, map[string]string{"pos": "in"})
//	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
//		t.Error(res.StatusCode, res.Body)
//	}
//}

func TestHandlePost(t *testing.T) {
	req := sam.NewRequest(http.MethodPost, nil)
	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

//func TestHandlePut(t *testing.T) {
//	req := sam.NewRequest(http.MethodPut, nil)
//	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
//		t.Error(res.StatusCode, res.Body)
//	}
//}
//
//func TestHandlePatch(t *testing.T) {
//	req := sam.NewRequest(http.MethodPatch, map[string]string{"id": "302191798223982"})
//	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
//		t.Error(res.StatusCode, res.Body)
//	} else {
//		pretty.Print(res.Body)
//	}
//	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
//		t.Error(res.StatusCode, res.Body)
//	} else {
//		pretty.Print(res.Body)
//	}
//}
