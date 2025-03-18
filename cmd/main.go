package main

import (
	"context"
	"fmt"
	"net/http"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/auth"
	"github.com/celestialdragonfly/betterreads/internal/env"
	"github.com/celestialdragonfly/betterreads/internal/log"
	"github.com/celestialdragonfly/betterreads/internal/middleware"
	"github.com/celestialdragonfly/betterreads/internal/mongo"
	"github.com/celestialdragonfly/betterreads/internal/server"
)

var (
	Host                   = env.GetDefault("BETTERREADS_HOST", "0.0.0.0")
	Port                   = env.GetIntDefault("BETTERREADS_PORT", 8080) //nolint: mnd // ignore magic numbers
	FirebaseServiceAccount = env.GetDefault("FIREBASE_SERVICE_ACCOUNT", "./secrets/firebase-serviceaccount.json")
	MongoUsername          = env.GetDefault("MONGO_USERNAME", "admin")
	MongoPassword          = env.GetDefault("MONGO_PASSWORD", "mangotango")
)

func main() {
	ctx := context.TODO()

	authClient, err := auth.NewFirebaseAuth(ctx, auth.Config{FirebaseServiceAccount: FirebaseServiceAccount})
	if err != nil {
		panic(fmt.Errorf("unable to start auth client %w", err))
	}

	mongoClient, err := mongo.NewMongoClient(ctx, MongoUsername, MongoPassword)
	if err != nil {
		panic(fmt.Errorf("unable to connect to mongo client %w", err))
	}

	server := server.NewServer(&server.Config{
		MongoClient: mongoClient,
	})

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

	log.Info(fmt.Sprintf("Server starting on port %d", Port))
	// Optional: close on shutdown
	defer func() {
		if err := mongoClient.DB.Disconnect(ctx); err != nil {
			panic(fmt.Sprintf("Error disconnecting: %v", err))
		}
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("Server failed to start: %v", err))
	}
}
