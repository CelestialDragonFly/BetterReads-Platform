FROM golang:1.24.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./cmd/binary ./cmd

FROM alpine:3.21.3 AS app-base
WORKDIR /app

FROM app-base AS betterreads
COPY --from=builder /app/cmd/binary betterreads
EXPOSE 8080
CMD ["./betterreads"]
