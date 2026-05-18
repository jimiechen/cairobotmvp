package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type HelloResponse struct {
	Result    Result `json:"result"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type Result struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	resp := HelloResponse{
		Result: Result{
			Code:    10200,
			Message: "success",
		},
		Message:   "Hello, World!",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/hello", helloHandler)
	http.ListenAndServe(":8080", nil)
}
