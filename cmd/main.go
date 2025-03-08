package main

import (
	"fmt"
	"log"
	"net/http"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/auth"
	"github.com/celestialdragonfly/betterreads/internal/env"
	"github.com/celestialdragonfly/betterreads/internal/middleware"
	"github.com/celestialdragonfly/betterreads/internal/server"
)

var (
	Host                   = env.GetDefault("BETTER_READS_HOST", "0.0.0.0")
	Port                   = env.GetIntDefault("BETTER_READS_PORT", 8080) //nolint: mnd // ignore magic numbers
	FirebaseServiceAccount = env.GetDefault("FIREBASE_SERVICE_ACCOUNT", "./secrets/firebase-serviceaccount.json")
)

func main() {
	server := server.NewServer(&server.Config{})

	authClient, err := auth.NewFirebaseAuth(auth.Config{FirebaseServiceAccount: FirebaseServiceAccount})
	if err != nil {
		panic(fmt.Errorf("unable to start auth client %w", err))
	}

	strictHandler := betterreads.NewStrictHandler(
		server,
		[]betterreads.StrictMiddlewareFunc{
			middleware.Authentication(authClient),
			middleware.Logging(),
		},
	)
	httpHandler := betterreads.Handler(strictHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", Host, Port),
		Handler: httpHandler,
	}

	log.Printf("Server starting on port %d", Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
