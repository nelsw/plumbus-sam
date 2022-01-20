package main

import (
	"context"
	"net/http"
	"plumbus/pkg/sam"
	"plumbus/pkg/util/pretty"
	"testing"
)

var ctx = context.TODO()

func TestHandleGetAll(t *testing.T) {
	req := sam.NewRequest(http.MethodGet, map[string]string{"pos": "all"})
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestHandleGetIn(t *testing.T) {
	req := sam.NewRequest(http.MethodGet, map[string]string{"pos": "in"})
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestHandleGetEx(t *testing.T) {
	req := sam.NewRequest(http.MethodGet, map[string]string{"pos": "ex"})
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestHandlePut(t *testing.T) {
	req := sam.NewRequest(http.MethodPut, nil)
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	}
}

func TestHandlePatch(t *testing.T) {
	req := sam.NewRequest(http.MethodPatch, map[string]string{"id": "302191798223982"})
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	} else {
		pretty.PrintJson(res.Body)
	}
	if res, _ := handle(ctx, req); res.StatusCode != http.StatusOK {
		t.Error(res.StatusCode, res.Body)
	} else {
		pretty.PrintJson(res.Body)
	}
}
