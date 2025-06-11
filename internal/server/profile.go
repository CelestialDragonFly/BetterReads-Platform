package server

import (
	"context"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

// (DELETE /api/v1/profile).
func (s *Server) DeleteUserProfile(ctx context.Context, request betterreads.DeleteUserProfileRequestObject) (betterreads.DeleteUserProfileResponseObject, error) {
	return nil, nil
}

// (GET /api/v1/profile).
func (s *Server) GetCurrentUserProfile(ctx context.Context, request betterreads.GetCurrentUserProfileRequestObject) (betterreads.GetCurrentUserProfileResponseObject, error) {
	return nil, nil
}

// (PATCH /api/v1/profile).
func (s *Server) UpdateUserProfile(ctx context.Context, request betterreads.UpdateUserProfileRequestObject) (betterreads.UpdateUserProfileResponseObject, error) {
	return nil, nil
}

// (POST /api/v1/profile).
func (s *Server) CreateUserProfile(ctx context.Context, request betterreads.CreateUserProfileRequestObject) (betterreads.CreateUserProfileResponseObject, error) {
	return nil, nil
}
