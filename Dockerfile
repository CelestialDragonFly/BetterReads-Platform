FROM --platform=$BUILDPLATFORM golang:1.26.0 AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o ./cmd/binary ./cmd

FROM alpine:3.21.3 AS app-base
WORKDIR /app
RUN apk add --no-cache ca-certificates
FROM app-base AS betterreads
COPY --from=builder /app/cmd/binary /app/betterreads
RUN chmod +x /app/betterreads
ARG FIREBASE_CONFIG
ENV FIREBASE_SERVICE_ACCOUNT=/app/firebase-serviceaccount.json
RUN echo "$FIREBASE_CONFIG" > /app/firebase-serviceaccount.json

EXPOSE 8080

CMD ["/app/betterreads"]
