#!/usr/bin/env sh
# Regenerates Go code from .proto files into backend/gen.
# Requires protoc, protoc-gen-go and protoc-gen-go-grpc on PATH.
set -e
cd "$(dirname "$0")/.."

MODULE=github.com/ArtemYarin/pinterest-clone-api

protoc -I proto \
  --go_out=. --go_opt=module=$MODULE \
  --go-grpc_out=. --go-grpc_opt=module=$MODULE \
  proto/likes/v1/likes.proto \
  proto/pin/v1/pin.proto
