package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloEndpointReturns200(t *testing.T) {
	req, err := http.NewRequest("GET", "/hello", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(helloHandler)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", rr.Code)
	}
}

func TestHelloEndpointReturnsJSON(t *testing.T) {
	req, err := http.NewRequest("GET", "/hello", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(helloHandler)

	handler.ServeHTTP(rr, req)

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("期望 Content-Type 为 application/json，实际 %s", contentType)
	}
}

func TestHelloEndpointContainsMessage(t *testing.T) {
	req, err := http.NewRequest("GET", "/hello", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(helloHandler)

	handler.ServeHTTP(rr, req)

	body := rr.Body.String()
	if len(body) == 0 || body == "" {
		t.Error("响应体为空")
	}

	if len(body) < 10 {
		t.Errorf("响应体过短: %s", body)
	}
}
