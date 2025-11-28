package server

import (
	"context"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

// (GET /api/v1/feed).
// GetPersonalizedFeed implements betterreads.BetterReadsServiceServer
func (s *Server) GetPersonalizedFeed(ctx context.Context, request *betterreads.GetPersonalizedFeedRequest) (*betterreads.GetPersonalizedFeedResponse, error) {
	return nil, nil
}

// GetUserFeed implements betterreads.BetterReadsServiceServer
func (s *Server) GetUserFeed(ctx context.Context, request *betterreads.GetUserFeedRequest) (*betterreads.GetUserFeedResponse, error) {
	return nil, nil
}

// CreatePost implements betterreads.BetterReadsServiceServer
func (s *Server) CreatePost(ctx context.Context, request *betterreads.CreatePostRequest) (*betterreads.CreatePostResponse, error) {
	return nil, nil
}

// DeletePost implements betterreads.BetterReadsServiceServer
func (s *Server) DeletePost(ctx context.Context, request *betterreads.DeletePostRequest) (*betterreads.DeletePostResponse, error) {
	return nil, nil
}

// UpdatePost implements betterreads.BetterReadsServiceServer
func (s *Server) UpdatePost(ctx context.Context, request *betterreads.UpdatePostRequest) (*betterreads.UpdatePostResponse, error) {
	return nil, nil
}

// GetCommentsForPost implements betterreads.BetterReadsServiceServer
func (s *Server) GetCommentsForPost(ctx context.Context, request *betterreads.GetCommentsForPostRequest) (*betterreads.GetCommentsForPostResponse, error) {
	return nil, nil
}

// AddComment implements betterreads.BetterReadsServiceServer
func (s *Server) AddComment(ctx context.Context, request *betterreads.AddCommentRequest) (*betterreads.AddCommentResponse, error) {
	return nil, nil
}

// DeleteComment implements betterreads.BetterReadsServiceServer
func (s *Server) DeleteComment(ctx context.Context, request *betterreads.DeleteCommentRequest) (*betterreads.DeleteCommentResponse, error) {
	return nil, nil
}

// LikePost implements betterreads.BetterReadsServiceServer
func (s *Server) LikePost(ctx context.Context, request *betterreads.LikePostRequest) (*betterreads.LikePostResponse, error) {
	return nil, nil
}

// UnlikePost implements betterreads.BetterReadsServiceServer
func (s *Server) UnlikePost(ctx context.Context, request *betterreads.UnlikePostRequest) (*betterreads.UnlikePostResponse, error) {
	return nil, nil
}
