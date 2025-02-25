package betterreads

//go:generate rm ./generated/betterreads_gen.go
//go:generate mkdir -p ./tmp
//go:generate curl -o ./tmp/betterreads.yaml https://raw.githubusercontent.com/CelestialDragonFly/BetterReads-OpenAPI/refs/heads/main/betterreads.yaml
//go:generate go run -mod=mod github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -generate types,strict-server,std-http-server -package betterreads -o ./generated/betterreads_gen.go ./tmp/betterreads.yaml
//go:generate rm -rf ./tmp
