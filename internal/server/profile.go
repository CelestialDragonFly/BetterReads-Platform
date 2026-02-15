package server

import (
	"context"
	"errors"
	"net/mail"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/celestialdragonfly/betterreads/internal/headers"
	"github.com/celestialdragonfly/betterreads/internal/logger"
	"github.com/celestialdragonfly/betterreads/internal/postgres"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DeleteUserProfile implements betterreads.BetterReadsServiceServer.
func (s *Server) DeleteUserProfile(ctx context.Context, _ *betterreads.DeleteUserProfileRequest) (*betterreads.DeleteUserProfileResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unable to retrieve user_id from context")
	}

	if err := s.DB.ProfileDelete(ctx, userID); err != nil {
		switch {
		case errors.Is(err, postgres.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, "failed to delete profile")
		}
	}

	return &betterreads.DeleteUserProfileResponse{}, nil
}

// GetCurrentUserProfile implements betterreads.BetterReadsServiceServer.
func (s *Server) GetCurrentUserProfile(ctx context.Context, _ *betterreads.GetCurrentUserProfileRequest) (*betterreads.GetCurrentUserProfileResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unable to retrieve user_id from context")
	}

	profile, err := s.DB.ProfileGet(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, "failed to get profile")
		}
	}

	return &betterreads.GetCurrentUserProfileResponse{
		CreatedAt:       timestamppb.New(profile.GetCreatedAt()),
		Email:           profile.GetEmail(),
		FirstName:       profile.GetFirstName(),
		Id:              profile.GetID(),
		LastName:        profile.GetLastName(),
		ProfilePhotoUrl: profile.GetProfilePhotoURL(),
		Username:        profile.GetUsername(),
	}, nil
}

// UpdateUserProfile implements betterreads.BetterReadsServiceServer.
func (s *Server) UpdateUserProfile(ctx context.Context, request *betterreads.UpdateUserProfileRequest) (*betterreads.UpdateUserProfileResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unable to retrieve user_id from context")
	}

	update := data.User{
		Username:        request.Username,
		FirstName:       request.FirstName,
		LastName:        request.LastName,
		Email:           request.Email,
		ProfilePhotoURL: request.ProfilePhotoUrl,
	}

	updatedUser, err := s.DB.ProfileUpdate(ctx, userID, &update)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		case errors.Is(err, postgres.ErrUserNameExists):
			return nil, status.Error(codes.AlreadyExists, "username already exists")
		case errors.Is(err, postgres.ErrEmailExists):
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		default:
			return nil, status.Error(codes.Internal, "failed to update profile")
		}
	}

	return &betterreads.UpdateUserProfileResponse{
		CreatedAt:       timestamppb.New(updatedUser.GetCreatedAt()),
		Email:           updatedUser.GetEmail(),
		FirstName:       updatedUser.GetFirstName(),
		Id:              updatedUser.GetID(),
		LastName:        updatedUser.GetLastName(),
		ProfilePhotoUrl: updatedUser.GetProfilePhotoURL(),
		Username:        updatedUser.GetUsername(),
	}, nil
}

// CreateUserProfile implements betterreads.BetterReadsServiceServer.
func (s *Server) CreateUserProfile(ctx context.Context, request *betterreads.CreateUserProfileRequest) (*betterreads.CreateUserProfileResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unable to retrieve user_id from context")
	}

	newProfile := data.User{
		ID:              userID,
		Username:        request.Username,
		FirstName:       request.FirstName,
		LastName:        request.LastName,
		Email:           request.Email,
		ProfilePhotoURL: request.ProfilePhotoUrl,
	}

	if len(newProfile.GetUsername()) < 3 { //nolint: mnd // minimum username length
		return nil, status.Error(codes.InvalidArgument, "username must be at least 3 characters")
	}

	if !isValidEmail(newProfile.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument, "invalid email address")
	}

	createdProfile, err := s.DB.ProfileCreate(ctx, &newProfile)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrUserNameExists):
			return nil, status.Error(codes.AlreadyExists, "username already exists")
		case errors.Is(err, postgres.ErrEmailExists):
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		default:
			return nil, status.Error(codes.Internal, "failed to create profile")
		}
	}

	return &betterreads.CreateUserProfileResponse{
		CreatedAt:       timestamppb.New(createdProfile.GetCreatedAt()),
		Email:           createdProfile.GetEmail(),
		FirstName:       createdProfile.GetFirstName(),
		Id:              createdProfile.GetID(),
		LastName:        createdProfile.GetLastName(),
		ProfilePhotoUrl: createdProfile.GetProfilePhotoURL(),
		Username:        createdProfile.GetUsername(),
	}, nil
}

// isValidEmail checks whether the given string is a valid email address.
func isValidEmail(email string) bool {
	if _, err := mail.ParseAddress(email); err != nil {
		logger.Error("Invalid email address", "email", email, "error", err)
		return false
	}
	return true
}
