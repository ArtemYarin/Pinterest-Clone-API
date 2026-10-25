package postgres

import (
	"net"
	"net/url"
	"os"
)

// DSNFromEnv builds a connection URL from POSTGRES_* env vars.
// Credentials are URL-escaped so passwords with @, :, / or # work.
func DSNFromEnv() string {
	dsn := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD")),
		Host:     net.JoinHostPort(os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT")),
		Path:     "/" + os.Getenv("POSTGRES_DB"),
		RawQuery: "sslmode=disable",
	}
	return dsn.String()
}
