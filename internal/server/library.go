package server

import (
	"context"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

// RemoveLibraryBook
func (s *Server) RemoveLibraryBook(
	ctx context.Context,
	request *betterreads.RemoveLibraryBookRequest,
) (*betterreads.RemoveLibraryBookResponse, error) {
	return nil, nil
}

// UpdateLibraryBook
func (s *Server) UpdateLibraryBook(
	ctx context.Context,
	request *betterreads.UpdateLibraryBookRequest,
) (*betterreads.UpdateLibraryBookResponse, error) {
	return &betterreads.UpdateLibraryBookResponse{}, nil
}

// GetUserLibrary
func (s *Server) GetUserLibrary(
	ctx context.Context,
	request *betterreads.GetUserLibraryRequest,
) (*betterreads.GetUserLibraryResponse, error) {
	return nil, nil
}
