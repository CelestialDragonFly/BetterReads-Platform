package betterreads

//go:generate ./generate.sh
//go:generate go run -mod=mod go.uber.org/mock/mockgen@latest -package=mocks -source=./internal/postgres/postgres.go -destination=./internal/postgres/mocks/postgres.go
