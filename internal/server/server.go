package server

import (
	"fmt"

	firebase "firebase.google.com/go/auth"
	"github.com/celestialdragonfly/betterreads-platform/internal/dependency/auth"
	iError "github.com/celestialdragonfly/betterreads-platform/internal/package/errors"
)

type Config struct {
	FirebaseJWTFilePath string
	GoogleBooksAPIKey   string
	NYTBooksAPIKey      string
}

type Server struct {
	Firebase          *firebase.Client
	GoogleBooksAPIKey string
	NYTBooksAPIKey    string
}

var errStartingServer = fmt.Errorf("unable to start better reads server")

func NewServer(cfg *Config) (*Server, error) {
	firebaseAuth, err := auth.NewAuth(cfg.FirebaseJWTFilePath)
	if err != nil {
		return nil, iError.WrapError(errStartingServer, err)
	}
	svr := Server{
		Firebase: firebaseAuth,
	}
	return &svr, nil
}
