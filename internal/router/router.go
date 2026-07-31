package router

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func SetupRouter() chi.Router {
	r := chi.NewRouter()

	r.HandleFunc("/pin*", proxyToPin)   // Pin proxy
	r.HandleFunc("/auth*", proxyToAuth) // Auth proxy

	return r
}

func proxyToAuth(w http.ResponseWriter, r *http.Request) {
	trimmedPath := strings.TrimPrefix(r.URL.Path, "/auth") // /auth/login -> /login
	authURL := "http://localhost:8081" + trimmedPath       // http://localhost:8081/login

	body, _ := io.ReadAll(r.Body)

	req, _ := http.NewRequest(r.Method, authURL, bytes.NewBuffer(body))
	req.Header = r.Header.Clone()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Auth service unavailable: %v", err), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func proxyToPin(w http.ResponseWriter, r *http.Request) {
	trimmedPath := strings.TrimPrefix(r.URL.Path, "/pin") // /pin/id -> /id
	pinURL := "http://localhost:8082" + trimmedPath       // http://localhost:8082/id

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read body: %v", err), http.StatusServiceUnavailable)
		return
	}

	req, err := http.NewRequest(r.Method, pinURL, bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate request: %v", err), http.StatusServiceUnavailable)
		return
	}
	req.Header = r.Header.Clone()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Pin service unavailable: %v", err), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to copy response body: %v", err), http.StatusServiceUnavailable)
		return
	}
}
