package main

import (
	"context"
	"net/http"
	"plumbus/pkg/sam"
	"testing"
)

var ctx = context.TODO()

func TestHandleGetArr(t *testing.T) {
	par := map[string]string{"form": "arr"}
	req := sam.NewRequest(http.MethodGet, par)
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestHandleGetMap(t *testing.T) {
	par := map[string]string{"form": "pam"}
	req := sam.NewRequest(http.MethodGet, par)
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestHandlePut(t *testing.T) {
	par := map[string]string{"id": "wat"}
	req := sam.NewRequest(http.MethodPut, par)
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestHandleDel(t *testing.T) {
	par := map[string]string{"id": "wat"}
	req := sam.NewRequest(http.MethodDelete, par)
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}
