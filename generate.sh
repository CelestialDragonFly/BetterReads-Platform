#!/bin/bash

function generate_open_api() {
	rm ./generated/betterreads_gen.go
	mkdir -p ./tmp
	curl -o ./tmp/betterreads.yaml https://raw.githubusercontent.com/CelestialDragonFly/BetterReads-OpenAPI/refs/heads/main/betterreads.yaml
	go run -mod=mod github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -generate types,strict-server,std-http-server -package betterreads -o ./generated/betterreads_gen.go ./tmp/betterreads.yaml
	rm -rf ./tmp
}

generate_open_api
