package router

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/ArtemYarin/pinterest-clone-api/pkg/middleware"
	"github.com/go-chi/chi/v5"
)

func SetupRouter(rateLimiter *middleware.IPRateLimiter) chi.Router {
	r := chi.NewRouter()

	r.Use(rateLimiter.RateLimitingMiddleware)

	r.HandleFunc("/auth*", proxyToAuth)
	r.HandleFunc("/pin*", proxyToPin)

	return r
}

func setupAuthProxy() *httputil.ReverseProxy {
	authUrl, _ := url.Parse("http://localhost:8081")

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
	pinURL, _ := url.Parse("http://localhost:8082")

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

var authProxy = setupAuthProxy()
var pinProxy = setupPinProxy()

func proxyToAuth(w http.ResponseWriter, r *http.Request) {
	authProxy.ServeHTTP(w, r)
}

func proxyToPin(w http.ResponseWriter, r *http.Request) {
	pinProxy.ServeHTTP(w, r)
}
