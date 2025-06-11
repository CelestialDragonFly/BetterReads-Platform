package server

import (
	"context"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

// (DELETE /api/v1/library).
func (s *Server) RemoveLibraryBook(ctx context.Context, request betterreads.RemoveLibraryBookRequestObject) (betterreads.RemoveLibraryBookResponseObject, error) {
	return nil, nil
}

// (PUT /api/v1/library).
func (s *Server) UpdateLibraryBook(ctx context.Context, request betterreads.UpdateLibraryBookRequestObject) (betterreads.UpdateLibraryBookResponseObject, error) {
	return nil, nil
}

// (GET /api/v1/library/{user_id}).
func (s *Server) GetUserLibrary(ctx context.Context, request betterreads.GetUserLibraryRequestObject) (betterreads.GetUserLibraryResponseObject, error) {
	return nil, nil
}
