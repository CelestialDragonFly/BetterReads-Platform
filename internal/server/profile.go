package server

import (
	"context"
	"errors"
	"net/mail"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/celestialdragonfly/betterreads/internal/headers"
)

// DeleteUserProfile implements betterreads.BetterReadsServiceServer
func (s *Server) DeleteUserProfile(
	ctx context.Context,
	request *betterreads.DeleteUserProfileRequest,
) (*betterreads.DeleteUserProfileResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, errors.New("unable to retrieve user_id from context")
	}

	if err := s.DB.ProfileDelete(ctx, userID); err != nil {
		return nil, err
	}

	return &betterreads.DeleteUserProfileResponse{}, nil
}

// GetCurrentUserProfile implements betterreads.BetterReadsServiceServer
func (s *Server) GetCurrentUserProfile(
	ctx context.Context,
	request *betterreads.GetCurrentUserProfileRequest,
) (*betterreads.GetProfileResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, errors.New("unable to retrieve user_id from context")
	}

	profile, err := s.DB.ProfileGet(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &betterreads.GetProfileResponse{
		CreatedAt:    profile.GetCreatedAt(),
		Email:        profile.GetEmail(),
		FirstName:    profile.GetFirstName(),
		Id:           profile.GetID(),
		LastName:     profile.GetLastName(),
		ProfilePhoto: profile.GetProfilePhoto(),
		Username:     profile.GetUsername(),
	}, nil
}

// UpdateUserProfile implements betterreads.BetterReadsServiceServer
func (s *Server) UpdateUserProfile(
	ctx context.Context,
	request *betterreads.UpdateUserProfileRequest,
) (*betterreads.UpdateProfileResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, errors.New("unable to retrieve user_id from context")
	}

	update := data.User{
		Username:     &request.Username,
		FirstName:    &request.FirstName,
		LastName:     &request.LastName,
		Email:        &request.Email,
		ProfilePhoto: &request.ProfilePhoto,
	}

	updatedUser, err := s.DB.ProfileUpdate(ctx, userID, &update)
	if err != nil {
		return nil, err
	}

	return &betterreads.UpdateProfileResponse{
		CreatedAt:    updatedUser.GetCreatedAt(),
		Email:        updatedUser.GetEmail(),
		FirstName:    updatedUser.GetFirstName(),
		Id:           updatedUser.GetID(),
		LastName:     updatedUser.GetLastName(),
		ProfilePhoto: updatedUser.GetProfilePhoto(),
		Username:     updatedUser.GetUsername(),
	}, nil
}

// CreateUserProfile implements betterreads.BetterReadsServiceServer
func (s *Server) CreateUserProfile(
	ctx context.Context,
	request *betterreads.CreateUserProfileRequest,
) (*betterreads.CreateProfileResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, errors.New("unable to retrieve user_id from context")
	}

	newProfile := data.User{
		ID:           &userID,
		Username:     &request.Username,
		FirstName:    &request.FirstName,
		LastName:     &request.LastName,
		Email:        &request.Email,
		ProfilePhoto: &request.ProfilePhoto,
	}

	if len(newProfile.GetUsername()) < 3 {
		return nil, errors.New("username must be at least 3 characters")
	}

	if !isValidEmail(newProfile.GetEmail()) {
		return nil, errors.New("invalid email address")
	}

	createdProfile, err := s.DB.ProfileCreate(ctx, &newProfile)
	if err != nil {
		return nil, err
	}

	return &betterreads.CreateProfileResponse{
		CreatedAt:    createdProfile.GetCreatedAt(),
		Email:        createdProfile.GetEmail(),
		FirstName:    createdProfile.GetFirstName(),
		Id:           createdProfile.GetID(),
		LastName:     createdProfile.GetLastName(),
		ProfilePhoto: createdProfile.GetProfilePhoto(),
		Username:     createdProfile.GetUsername(),
	}, nil
}

// isValidEmail checks whether the given string is a valid email address.
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
