package router

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ArtemYarin/pinterest-clone-api/internal/app/pin"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRouter(pinHandler pin.PinHandler, pinPool *pgxpool.Pool) chi.Router {
	r := chi.NewRouter()

	r.Mount("/pins", pin.PinRouter(&pinHandler, pinPool))
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
