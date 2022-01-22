package main

import (
	"net/http"
	"plumbus/pkg/sam"
	"plumbus/test"
	"testing"
)

func TestHandlePut(t *testing.T) {

	req := sam.NewRequest(http.MethodPut, nil)
	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}

}

func TestHandleGet(t *testing.T) {

	req := sam.NewRequest(http.MethodGet, nil)
	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}

}
