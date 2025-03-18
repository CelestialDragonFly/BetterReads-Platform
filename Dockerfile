FROM golang:1.24.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/binary ./cmd

FROM alpine:3.21.3 AS app-base
WORKDIR /app
RUN apk add --no-cache ca-certificates
FROM app-base AS betterreads
COPY --from=builder /app/cmd/binary /app/betterreads
RUN chmod +x /app/betterreads
ARG FIREBASE_CONFIG
ENV FIREBASE_SERVICE_ACCOUNT=/app/firebase-serviceaccount.json

EXPOSE 8080

# Run the binary
CMD ["/app/betterreads"]
