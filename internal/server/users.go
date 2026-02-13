package server

import (
	"context"
	"errors"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/headers"
	"github.com/celestialdragonfly/betterreads/internal/postgres"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// (GET /api/v1/users/{user_id}).
// GetUserById implements betterreads.BetterReadsServiceServer.
func (s *Server) GetUserById(ctx context.Context, request *betterreads.GetUserByIdRequest) (*betterreads.GetUserByIdResponse, error) { //nolint:revive // generated method name
	user, err := s.DB.GetUserByID(ctx, request.GetUserId())
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, "failed to get user")
		}
	}
	return &betterreads.GetUserByIdResponse{
		Id:              user.ID,
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		ProfilePhotoUrl: user.ProfilePhotoURL,
	}, nil
}

// FollowUser implements betterreads.BetterReadsServiceServer.
func (s *Server) FollowUser(ctx context.Context, request *betterreads.FollowUserRequest) (*betterreads.FollowUserResponse, error) {
	followeeID := request.GetUserId()
	followerID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unable to retrieve user_id from context")
	}

	err := s.DB.FollowUser(ctx, followerID, followeeID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrSelfFollow):
			return nil, status.Error(codes.InvalidArgument, "cannot follow yourself")
		default:
			return nil, status.Error(codes.Internal, "failed to follow user")
		}
	}
	return &betterreads.FollowUserResponse{}, nil
}

// UnfollowUser implements betterreads.BetterReadsServiceServer.
func (s *Server) UnfollowUser(ctx context.Context, request *betterreads.UnfollowUserRequest) (*betterreads.UnfollowUserResponse, error) {
	followeeID := request.GetUserId()
	followerID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unable to retrieve user_id from context")
	}

	err := s.DB.UnfollowUser(ctx, followerID, followeeID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrSelfFollow):
			return nil, status.Error(codes.InvalidArgument, "cannot unfollow yourself")
		default:
			return nil, status.Error(codes.Internal, "failed to unfollow user")
		}
	}
	return &betterreads.UnfollowUserResponse{}, nil
}
