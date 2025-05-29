package openlibrary

import (
	"context"
	"fmt"
	"net/url"

	library "github.com/celestialdragonfly/betterreads/internal/openlibrary/contracts"
)

//go:generate go run -mod=mod github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest --generate=client,types -package=openlibrary -o ./contracts/openlibrary.gen.go ./contracts/openlibrary.yaml

// ClientInterface defines the interface for interacting with the OpenLibrary API.
type ClientInterface interface {
	// SearchBooks searches for books in the OpenLibrary based on a query string
	SearchBooks(ctx context.Context, query string, title, author, subject *string) (*SearchBooksResponse, error)
}

type Client struct {
	Client *library.Client
}

var _ ClientInterface = (*Client)(nil)

func NewClient(host string) (*Client, error) {
	// Validate URL format before creating client
	_, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("invalid host URL: %w", err)
	}

	openLibraryClient, err := library.NewClient(host)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: openLibraryClient,
	}, nil
}

func getFirstValue[V interface{ int | string }](v []V) V {
	if len(v) == 0 {
		var defaultValue V
		return defaultValue
	}
	return v[0]
}
