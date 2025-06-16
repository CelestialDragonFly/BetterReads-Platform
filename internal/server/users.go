package server

import (
	"context"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

// (GET /api/v1/users/{user_id}).
func (s *Server) GetUserById(ctx context.Context, request betterreads.GetUserByIdRequestObject) (betterreads.GetUserByIdResponseObject, error) {
	return nil, nil
}

// (POST /api/v1/users/{user_id}/follow).
func (s *Server) FollowUser(ctx context.Context, request betterreads.FollowUserRequestObject) (betterreads.FollowUserResponseObject, error) {
	return nil, nil
}

// (DELETE /api/v1/users/{user_id}/unfollow).
func (s *Server) UnfollowUser(ctx context.Context, request betterreads.UnfollowUserRequestObject) (betterreads.UnfollowUserResponseObject, error) {
	return nil, nil
}
