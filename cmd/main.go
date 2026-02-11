package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/auth"
	"github.com/celestialdragonfly/betterreads/internal/env"
	"github.com/celestialdragonfly/betterreads/internal/log"
	"github.com/celestialdragonfly/betterreads/internal/middleware"
	"github.com/celestialdragonfly/betterreads/internal/openlibrary"
	"github.com/celestialdragonfly/betterreads/internal/postgres"
	"github.com/celestialdragonfly/betterreads/internal/server"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Host                   = env.GetDefault("BETTERREADS_HOST", "0.0.0.0")
	Port                   = env.GetIntDefault("BETTERREADS_PORT", 8080) //nolint: mnd // ignore magic numbers
	GRPCPort               = env.GetIntDefault("BETTERREADS_GRPC_PORT", 9090)
	FirebaseServiceAccount = env.GetDefault("FIREBASE_SERVICE_ACCOUNT", "./secrets/firebase-serviceaccount.json")
	SQLURL                 = env.GetDefault("SQL_URL", "postgresql://admin:sqltango@localhost:5432/betterreads")
	OpenLibraryHost        = env.GetDefault("OPEN_LIBRARY_HOST", "https://openlibrary.org")
	timeout                = 5 * time.Second
	ReaderTimeout          = env.GetDurationDefault("BETTERREADS_READERTIMEOUT", timeout)
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	authClient, err := auth.NewFirebaseAuth(ctx, auth.Config{FirebaseServiceAccount: FirebaseServiceAccount})
	if err != nil {
		panic(fmt.Errorf("unable to start auth client %w", err))
	}

	sqlClient, err := postgres.NewClient(ctx, SQLURL)
	if err != nil {
		panic(fmt.Errorf("unable to connect to postgres client %w", err))
	}
	defer func() {
		if err := sqlClient.DB.Close(ctx); err != nil {
			log.Error(fmt.Sprintf("Error disconnecting: %v", err))
		}
	}()

	openLibraryClient, err := openlibrary.NewClient(OpenLibraryHost)
	if err != nil {
		panic(fmt.Errorf("unable to connect to open library %w", err))
	}

	srv := server.NewServer(&server.Config{
		SQLClient:   sqlClient,
		OpenLibrary: openLibraryClient,
	})

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", Host, GRPCPort))
	if err != nil {
		panic(fmt.Errorf("failed to listen: %w", err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GRPCAuthentication(authClient)),
	)
	betterreads.RegisterBetterReadsServiceServer(grpcServer, srv)

	go func() {
		log.Info(fmt.Sprintf("gRPC Server starting on port %d", GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			panic(fmt.Errorf("failed to serve gRPC: %w", err))
		}
	}()

	// Start gRPC-Gateway
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = betterreads.RegisterBetterReadsServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", GRPCPort), opts)
	if err != nil {
		panic(fmt.Errorf("failed to register gateway: %w", err))
	}

	httpSrv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", Host, Port),
		Handler:           mux,
		ReadHeaderTimeout: ReaderTimeout,
	}

	log.Info(fmt.Sprintf("HTTP Gateway starting on port %d", Port))
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("HTTP Server failed to start: %v", err))
	}
}
