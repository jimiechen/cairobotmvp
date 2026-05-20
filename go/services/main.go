package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type HelloWorldResponse struct {
	Result    Result `json:"result"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}

	message := "Hello, " + name + "!"
	resp := HelloWorldResponse{
		Result: Result{
			Code: 0,
			Msg:  "success",
		},
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/hello", helloHandler)
	http.ListenAndServe(":8080", nil)
}
