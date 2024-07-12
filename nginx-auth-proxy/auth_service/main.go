package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type TokenResponse struct {
	Token string `json:"token"`
}

func main() {
	http.HandleFunc("/auth", authHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func authHandler(w http.ResponseWriter, r *http.Request) {

	user := r.Header.Get("X-Auth-Request-User")
	if user == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	st2AuthURL := os.Getenv("ST2AUTH_URL")

	payload := map[string]string{
		"user":        user,
		"remote_addr": r.Header.Get("X-Real-IP"),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(st2AuthURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error calling StackStorm API: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("StackStorm API returned non-200 status: %d", resp.StatusCode)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Auth-Request-User", user)
	w.Header().Set("X-ST2-Token", tokenResp.Token)
	fmt.Fprintf(w, "OK")
}
