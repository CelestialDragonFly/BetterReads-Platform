package server

import (
	"context"
	"errors"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

// (GET /api/v1/feed).
// GetPersonalizedFeed implements betterreads.BetterReadsServiceServer.
func (s *Server) GetPersonalizedFeed(_ context.Context, _ *betterreads.GetPersonalizedFeedRequest) (*betterreads.GetPersonalizedFeedResponse, error) {
	return nil, errors.New("not implemented")
}

// GetUserFeed implements betterreads.BetterReadsServiceServer.
func (s *Server) GetUserFeed(_ context.Context, _ *betterreads.GetUserFeedRequest) (*betterreads.GetUserFeedResponse, error) {
	return nil, errors.New("not implemented")
}

// CreatePost implements betterreads.BetterReadsServiceServer.
func (s *Server) CreatePost(_ context.Context, _ *betterreads.CreatePostRequest) (*betterreads.CreatePostResponse, error) {
	return nil, errors.New("not implemented")
}

// DeletePost implements betterreads.BetterReadsServiceServer.
func (s *Server) DeletePost(_ context.Context, _ *betterreads.DeletePostRequest) (*betterreads.DeletePostResponse, error) {
	return nil, errors.New("not implemented")
}

// UpdatePost implements betterreads.BetterReadsServiceServer.
func (s *Server) UpdatePost(_ context.Context, _ *betterreads.UpdatePostRequest) (*betterreads.UpdatePostResponse, error) {
	return nil, errors.New("not implemented")
}

// GetCommentsForPost implements betterreads.BetterReadsServiceServer.
func (s *Server) GetCommentsForPost(_ context.Context, _ *betterreads.GetCommentsForPostRequest) (*betterreads.GetCommentsForPostResponse, error) {
	return nil, errors.New("not implemented")
}

// AddComment implements betterreads.BetterReadsServiceServer.
func (s *Server) AddComment(_ context.Context, _ *betterreads.AddCommentRequest) (*betterreads.AddCommentResponse, error) {
	return nil, errors.New("not implemented")
}

// DeleteComment implements betterreads.BetterReadsServiceServer.
func (s *Server) DeleteComment(_ context.Context, _ *betterreads.DeleteCommentRequest) (*betterreads.DeleteCommentResponse, error) {
	return nil, errors.New("not implemented")
}

// LikePost implements betterreads.BetterReadsServiceServer.
func (s *Server) LikePost(_ context.Context, _ *betterreads.LikePostRequest) (*betterreads.LikePostResponse, error) {
	return nil, errors.New("not implemented")
}

// UnlikePost implements betterreads.BetterReadsServiceServer.
func (s *Server) UnlikePost(_ context.Context, _ *betterreads.UnlikePostRequest) (*betterreads.UnlikePostResponse, error) {
	return nil, errors.New("not implemented")
}
