package router

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func SetupRouter(rateLimiter *middleware.IPRateLimiter) chi.Router {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		// For dev/testing only:
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(rateLimiter.RateLimitingMiddleware)

	r.Get("/openapi.yaml", serveOpenAPISpec)
	r.Get("/docs", serveSwaggerUI)

	authProxy := setupAuthProxy()
	pinProxy := setupPinProxy()
	interactionProxy := setupInteractionProxy()

	r.HandleFunc("/auth*", func(w http.ResponseWriter, r *http.Request) {
		authProxy.ServeHTTP(w, r)
	})
	r.HandleFunc("/pin*", func(w http.ResponseWriter, r *http.Request) {
		pinProxy.ServeHTTP(w, r)
	})
	r.HandleFunc("/interaction*", func(w http.ResponseWriter, r *http.Request) {
		interactionProxy.ServeHTTP(w, r)
	})

	return r
}

func setupAuthProxy() *httputil.ReverseProxy {
	authUrl, _ := url.Parse(fmt.Sprintf("http://%s:%s", os.Getenv("AUTH_HOST"), os.Getenv("AUTH_PORT")))

	authProxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(authUrl)

			pr.Out.URL.Path = strings.TrimPrefix(pr.In.URL.Path, "/auth")
			if pr.Out.URL.Path == "" {
				pr.Out.URL.Path = "/"
			}

			pr.Out.Host = authUrl.Host

			pr.Out.Header.Set("X-Forwarded-Host", pr.In.Host)
			pr.Out.Header.Set("X-Forwarded-Proto", pr.In.URL.Scheme)
		},

		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 60,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	return authProxy
}

func setupPinProxy() *httputil.ReverseProxy {
	pinURL, _ := url.Parse(fmt.Sprintf("http://%s:%s", os.Getenv("PIN_HOST"), os.Getenv("PIN_PORT")))

	pinProxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(pinURL)

			pr.Out.URL.Path = strings.TrimPrefix(pr.In.URL.Path, "/pin")
			if pr.Out.URL.Path == "" {
				pr.Out.URL.Path = "/"
			}

			pr.Out.Host = pinURL.Host

			pr.Out.Header.Set("X-Forwarded-Host", pr.In.Host)
			pr.Out.Header.Set("X-Forwarded-Proto", pr.In.URL.Scheme)
		},

		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 60,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return pinProxy
}

func setupInteractionProxy() *httputil.ReverseProxy {
	interactionURL, _ := url.Parse(fmt.Sprintf("http://%s:%s", os.Getenv("INTERACTION_HOST"), os.Getenv("INTERACTION_PORT")))

	interactionProxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(interactionURL)

			pr.Out.URL.Path = strings.TrimPrefix(pr.In.URL.Path, "/interaction")
			if pr.Out.URL.Path == "" {
				pr.Out.URL.Path = "/"
			}

			pr.Out.Host = interactionURL.Host

			pr.Out.Header.Set("X-Forwarded-Host", pr.In.Host)
			pr.Out.Header.Set("X-Forwarded-Proto", pr.In.URL.Scheme)
		},

		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 60,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return interactionProxy
}
