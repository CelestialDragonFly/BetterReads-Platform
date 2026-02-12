package server

import (
	"context"
	"errors"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

// RemoveLibraryBook.
func (s *Server) RemoveLibraryBook(_ context.Context, _ *betterreads.RemoveLibraryBookRequest) (*betterreads.RemoveLibraryBookResponse, error) {
	return nil, errors.New("not implemented")
}

// UpdateLibraryBook.
func (s *Server) UpdateLibraryBook(_ context.Context, _ *betterreads.UpdateLibraryBookRequest) (*betterreads.UpdateLibraryBookResponse, error) {
	return &betterreads.UpdateLibraryBookResponse{}, nil
}

// GetUserLibrary.
func (s *Server) GetUserLibrary(_ context.Context, _ *betterreads.GetUserLibraryRequest) (*betterreads.GetUserLibraryResponse, error) {
	return nil, errors.New("not implemented")
}
