package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// 测试 HelloWorld 接口响应
func TestHelloWorldEndpoint(t *testing.T) {
	// 设置测试请求
	req, err := http.NewRequest("GET", "/hello?name=CaiRobot", nil)
	if err != nil {
		t.Fatal(err)
	}

	// 记录响应
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(helloHandler)

	// 处理请求
	handler.ServeHTTP(rr, req)

	// 检查状态码
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
