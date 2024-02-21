package main

import (
	"fmt"
	"net/http"

	"github.com/celestialdragonfly/betterreads-platform/internal/package/middleware"
	"github.com/celestialdragonfly/betterreads-platform/internal/server"
)

// TODO replace with flags
var (
	secretPath = "../secrets/betterreads-4e773-firebase-adminsdk-9g50q-580bddcbc7.json"
	host       = "localhost"
	port       = 8080
)

func main() {
	// TODO replace with structured logging
	fmt.Println("BetterReads Service")
	authServer, err := server.NewServer(&server.Config{
		FirebaseJWTFilePath: secretPath,
	})
	if err != nil {
		panic("unable to start BetterReads server")
	}

	mux := newMuxServer(authServer)
	// TODO replace with structured logging
	fmt.Printf("Starting service on host: %s | port: %d\n", host, port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), mux)
}

func newMuxServer(svr *server.Server) *http.ServeMux {
	unauthenticated := map[string]http.HandlerFunc{
		"GET /authuser": svr.AuthUser,
	}

	mux := http.NewServeMux()
	midware := []middleware.Middleware{middleware.Logger}
	for endpoint, f := range unauthenticated {
		mux.HandleFunc(endpoint, middleware.MultipleMiddleware(f, midware...))
	}

	return mux
}
