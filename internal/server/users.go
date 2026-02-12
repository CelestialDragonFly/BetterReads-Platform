package server

import (
	"context"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

// (GET /api/v1/users/{user_id}).
// GetUserById implements betterreads.BetterReadsServiceServer.
func (s *Server) GetUserById(_ context.Context, _ *betterreads.GetUserByIdRequest) (*betterreads.GetUserByIdResponse, error) { //nolint:revive // generated method name
	return &betterreads.GetUserByIdResponse{}, nil
}

// FollowUser implements betterreads.BetterReadsServiceServer.
func (s *Server) FollowUser(_ context.Context, _ *betterreads.FollowUserRequest) (*betterreads.FollowUserResponse, error) {
	return &betterreads.FollowUserResponse{}, nil
}

// UnfollowUser implements betterreads.BetterReadsServiceServer.
func (s *Server) UnfollowUser(_ context.Context, _ *betterreads.UnfollowUserRequest) (*betterreads.UnfollowUserResponse, error) {
	return &betterreads.UnfollowUserResponse{}, nil
}
