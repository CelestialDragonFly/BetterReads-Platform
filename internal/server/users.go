package server

import (
	"context"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

// (GET /api/v1/users/{user_id}).
// GetUserById implements betterreads.BetterReadsServiceServer
func (s *Server) GetUserById(ctx context.Context, request *betterreads.GetUserByIdRequest) (*betterreads.GetUserByIdResponse, error) {
	return &betterreads.GetUserByIdResponse{}, nil
}

// FollowUser implements betterreads.BetterReadsServiceServer
func (s *Server) FollowUser(ctx context.Context, request *betterreads.FollowUserRequest) (*betterreads.FollowUserResponse, error) {
	return &betterreads.FollowUserResponse{}, nil
}

// UnfollowUser implements betterreads.BetterReadsServiceServer
func (s *Server) UnfollowUser(ctx context.Context, request *betterreads.UnfollowUserRequest) (*betterreads.UnfollowUserResponse, error) {
	return &betterreads.UnfollowUserResponse{}, nil
}
