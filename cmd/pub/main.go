package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("[PUBLISHER]")
	for _ = range 10 {
		doReq()
		time.Sleep(1 * time.Second)
	}
}

func doReq() {
	req, err := http.NewRequest("POST", "http://localhost:8080/publish", bytes.NewBuffer([]byte("{\"Hello\": \"Friend\"}")))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status Code (Method 2):", resp.StatusCode)
}
