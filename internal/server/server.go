package server

import (
	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/openlibrary"
	"github.com/celestialdragonfly/betterreads/internal/postgres"
)

type Config struct {
	SQLClient   *postgres.Client
	OpenLibrary openlibrary.ClientInterface
}

type Server struct {
	betterreads.UnimplementedBetterReadsServiceServer
	DB          *postgres.Client
	OpenLibrary openlibrary.ClientInterface
}

var _ betterreads.BetterReadsServiceServer = (*Server)(nil)

// NewServer creates a new server instance with the provided configuration.
func NewServer(cfg *Config) *Server {
	return &Server{
		DB:          cfg.SQLClient,
		OpenLibrary: cfg.OpenLibrary,
	}
}

func getStringFromPointer(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
