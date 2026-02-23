#!/bin/bash
export DOCKER_HOST=unix:///var/run/docker.sock
DOCKER_CONFIG=$(mktemp -d)
export DOCKER_CONFIG
GOPATH=$(go env GOPATH)

function generate_grpc() {
	rm -rf ./generated/*

	# Build the builder image
	docker build -t betterreads-protoc -f Dockerfile.protoc .

	# Run generation
	docker run --rm \
		--volume "$(pwd):/workspace" \
		--volume "$GOPATH:/go" \
		-w /workspace \
		-e DOCKER_CONFIG=/tmp/docker-config \
		betterreads-protoc \
		-I/usr/local/include -Iproto \
		--go_out=generated --go_opt=paths=source_relative \
		--go-grpc_out=generated --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=generated --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=. --openapiv2_opt allow_merge=true,merge_file_name=betterreads \
		betterreads.proto
}

function generate_openapi_from_swagger() {
	docker run --rm -v "${PWD}":/usr/src/app mermade/swagger2openapi swagger2openapi --yaml --outfile /usr/src/app/betterreads.openapi.yaml /usr/src/app/betterreads.swagger.json
	sudo chown "$USER":"$USER" betterreads.openapi.yaml
}

generate_grpc
generate_openapi_from_swagger
