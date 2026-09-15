module github.com/ArtemYarin/pinterest-clone-api/services/interaction-service

go 1.26.1

require (
	github.com/ArtemYarin/pinterest-clone-api v0.0.0-20260901101134-4b8e8757b0c2
	github.com/alicebob/miniredis/v2 v2.39.0
	github.com/go-chi/chi/v5 v5.3.2
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.10.0
	github.com/joho/godotenv v1.5.1
	github.com/redis/go-redis/v9 v9.22.0
	github.com/stretchr/testify v1.12.1
	golang.org/x/sync v0.22.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace github.com/ArtemYarin/pinterest-clone-api => ../..
