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
	// Set up logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	http.HandleFunc("/auth", authHandler)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received auth request from %s", r.RemoteAddr)

	user := r.Header.Get("REMOTE_USER")
	if user == "" {
		log.Println("Unauthorized: X-Auth-Request-User header is empty")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("Authenticating user: %s", user)

	st2AuthURL := os.Getenv("ST2AUTH_URL")
	log.Printf("ST2AUTH_URL: %s", st2AuthURL)

	payload := map[string]string{
		"user":        user,
		"remote_addr": r.Header.Get("REMOTE_ADDR"),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling payload: %v", err)
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

	log.Printf("StackStorm API response status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		log.Printf("StackStorm API returned non-200 status: %d", resp.StatusCode)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Printf("Error unmarshalling response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Auth-Request-User", user)
	w.Header().Set("X-ST2-Token", tokenResp.Token)
	log.Printf("Authentication successful for user: %s", user)
	fmt.Fprintf(w, "OK")
}
