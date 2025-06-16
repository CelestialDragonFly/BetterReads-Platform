package server

import (
	"context"
	"errors"
	"net/mail"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/celestialdragonfly/betterreads/internal/headers"
	"github.com/celestialdragonfly/betterreads/internal/postgres"
	"github.com/google/uuid"
)

// (DELETE /api/v1/profile).
func (s *Server) DeleteUserProfile(ctx context.Context, request betterreads.DeleteUserProfileRequestObject) (betterreads.DeleteUserProfileResponseObject, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return betterreads.DeleteUserProfile500JSONResponse{
			Code: "INTERNAL_SERVER_ERROR",
			Details: &map[string]any{
				"error":        "unable to retrieve user_id from header",
				"reference_id": uuid.New(),
			},
			Message: "create user profile - internal server error",
		}, nil
	}

	if err := s.DB.ProfileDelete(ctx, userID); err != nil {
		return betterreads.DeleteUserProfile401JSONResponse{
			Code: "BAD_REQUEST",
			Details: &map[string]any{
				"error":        err.Error(),
				"reference_id": uuid.New(),
			},
			Message: "delete user profile - unable to delete user's profile",
		}, nil
	}

	return betterreads.DeleteUserProfile204Response{}, nil
}

// (GET /api/v1/profile).
func (s *Server) GetCurrentUserProfile(ctx context.Context, request betterreads.GetCurrentUserProfileRequestObject) (betterreads.GetCurrentUserProfileResponseObject, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return betterreads.GetCurrentUserProfile500JSONResponse{
			Code: "INTERNAL_SERVER_ERROR",
			Details: &map[string]any{
				"error":        "unable to retrieve user_id from header",
				"reference_id": uuid.New(),
			},
			Message: "get user profile - internal server error",
		}, nil
	}

	profile, err := s.DB.ProfileGet(ctx, userID)
	if err != nil {
		return betterreads.GetCurrentUserProfile500JSONResponse{
			Code: "NOT_FOUND",
			Details: &map[string]any{
				"error":        err.Error(),
				"reference_id": uuid.New(),
			},
			Message: "get user profile - not found",
		}, nil
	}

	return betterreads.GetCurrentUserProfile200JSONResponse{
		CreatedAt:    profile.GetCreatedAt(),
		Email:        profile.GetEmail(),
		FirstName:    profile.GetFirstName(),
		Id:           profile.GetID(),
		LastName:     profile.GetLastName(),
		ProfilePhoto: profile.GetProfilePhoto(),
		Username:     profile.GetUsername(),
	}, nil
}

// (PATCH /api/v1/profile).
func (s *Server) UpdateUserProfile(ctx context.Context, request betterreads.UpdateUserProfileRequestObject) (betterreads.UpdateUserProfileResponseObject, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return betterreads.UpdateUserProfile500JSONResponse{
			Code: "INTERNAL_SERVER_ERROR",
			Details: &map[string]any{
				"error":        "unable to retrieve user_id from header",
				"reference_id": uuid.New(),
			},
			Message: "update user profile - internal server error",
		}, nil
	}

	update := data.User{
		Username:     request.Body.Username,
		FirstName:    request.Body.FirstName,
		LastName:     request.Body.LastName,
		Email:        request.Body.Email,
		ProfilePhoto: request.Body.ProfilePhoto,
	}

	updatedUser, err := s.DB.ProfileUpdate(ctx, userID, &update)
	if err != nil {
		return betterreads.UpdateUserProfile400JSONResponse{
			Code: "BAD_REQUEST",
			Details: &map[string]any{
				"error":        err.Error(),
				"reference_id": uuid.New(),
			},
			Message: "update user profile - invalid request",
		}, nil
	}

	return betterreads.UpdateUserProfile200JSONResponse{
		CreatedAt:    updatedUser.GetCreatedAt(),
		Email:        updatedUser.GetEmail(),
		FirstName:    updatedUser.GetFirstName(),
		Id:           updatedUser.GetID(),
		LastName:     updatedUser.GetLastName(),
		ProfilePhoto: updatedUser.GetProfilePhoto(),
		Username:     updatedUser.GetUsername(),
	}, nil
}

// (POST /api/v1/profile).
func (s *Server) CreateUserProfile(ctx context.Context, request betterreads.CreateUserProfileRequestObject) (betterreads.CreateUserProfileResponseObject, error) {

	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return betterreads.CreateUserProfile500JSONResponse{
			Code: "INTERNAL_SERVER_ERROR",
			Details: &map[string]any{
				"error":        "unable to retrieve user_id from header",
				"reference_id": uuid.New(),
			},
			Message: "create user profile - internal server error",
		}, nil
	}

	newProfile := data.User{
		ID:           &userID,
		Username:     &request.Body.Username,
		FirstName:    &request.Body.FirstName,
		LastName:     &request.Body.LastName,
		Email:        &request.Body.Email,
		ProfilePhoto: &request.Body.ProfilePhoto,
	}

	if len(newProfile.GetUsername()) < 3 {
		return betterreads.CreateUserProfile400JSONResponse{
			Code: "VALIDATION_ERROR",
			Details: &map[string]any{
				"error":        "username must be at least 3 characters",
				"reference_id": uuid.New(),
			},
			Message: "create user profile - invalid request",
		}, nil
	}

	if isValidEmail(newProfile.GetEmail()) {
		return betterreads.CreateUserProfile400JSONResponse{
			Code: "VALIDATION_ERROR",
			Details: &map[string]any{
				"error":        "invalid email address",
				"reference_id": uuid.New(),
			},
			Message: "create user profile - invalid request",
		}, nil
	}

	createdProfile, err := s.DB.ProfileCreate(ctx, &newProfile)
	if err != nil {
		code := "BAD_REQUEST"
		if errors.Is(err, postgres.ErrEmailExists) {
			code = "BAD_REQUEST_DUPLICATE_EMAIL"
		} else if errors.Is(err, postgres.ErrUserNameExists) {
			code = "BAD_REQUEST_DUPLICATE_USER"
		}
		return betterreads.CreateUserProfile409JSONResponse{
			Code: code,
			Details: &map[string]any{
				"error":        err.Error(),
				"reference_id": uuid.New(),
			},
			Message: "create user profile - invalid request",
		}, nil
	}

	return betterreads.CreateUserProfile201JSONResponse{
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
