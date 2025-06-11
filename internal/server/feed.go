package server

import (
	"context"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

// (GET /api/v1/feed).
func (s *Server) GetPersonalizedFeed(ctx context.Context, request betterreads.GetPersonalizedFeedRequestObject) (betterreads.GetPersonalizedFeedResponseObject, error) {
	return nil, nil
}

// (GET /api/v1/feed/{user_id}).
func (s *Server) GetUserFeed(ctx context.Context, request betterreads.GetUserFeedRequestObject) (betterreads.GetUserFeedResponseObject, error) {
	return nil, nil
}

// (POST /api/v1/posts).
func (s *Server) CreatePost(ctx context.Context, request betterreads.CreatePostRequestObject) (betterreads.CreatePostResponseObject, error) {
	return nil, nil
}

// (DELETE /api/v1/posts/{post_id}).
func (s *Server) DeletePost(ctx context.Context, request betterreads.DeletePostRequestObject) (betterreads.DeletePostResponseObject, error) {
	return nil, nil
}

// (PUT /api/v1/posts/{post_id}).
func (s *Server) UpdatePost(ctx context.Context, request betterreads.UpdatePostRequestObject) (betterreads.UpdatePostResponseObject, error) {
	return nil, nil
}

// (GET /api/v1/posts/{post_id}/comments).
func (s *Server) GetCommentsForPost(ctx context.Context, request betterreads.GetCommentsForPostRequestObject) (betterreads.GetCommentsForPostResponseObject, error) {
	return nil, nil
}

// (POST /api/v1/posts/{post_id}/comments).
func (s *Server) AddComment(ctx context.Context, request betterreads.AddCommentRequestObject) (betterreads.AddCommentResponseObject, error) {
	return nil, nil
}

// (DELETE /api/v1/posts/{post_id}/comments/{comment_id}).
func (s *Server) DeleteComment(ctx context.Context, request betterreads.DeleteCommentRequestObject) (betterreads.DeleteCommentResponseObject, error) {
	return nil, nil
}

// (POST /api/v1/posts/{post_id}/like).
func (s *Server) LikePost(ctx context.Context, request betterreads.LikePostRequestObject) (betterreads.LikePostResponseObject, error) {
	return nil, nil
}

// (DELETE /api/v1/posts/{post_id}/unlike).
func (s *Server) UnlikePost(ctx context.Context, request betterreads.UnlikePostRequestObject) (betterreads.UnlikePostResponseObject, error) {
	return nil, nil
}
