package main

import (
	"net/http"
	"plumbus/pkg/sam"
	"plumbus/test"
	"testing"
)

func TestHandle(t *testing.T) {

	req := sam.NewRequest(http.MethodGet, nil)
	if res, _ := handle(test.CTX, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}

}
