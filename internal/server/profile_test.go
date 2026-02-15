package server

import (
	"context"
	"errors"
	"testing"
	"time"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/celestialdragonfly/betterreads/internal/headers"
	"github.com/celestialdragonfly/betterreads/internal/postgres"
	"github.com/celestialdragonfly/betterreads/internal/postgres/mocks"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper function to check error status code.
func checkErrorCode(t *testing.T, err error, wantCode codes.Code, methodName string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s() error = nil, wantErr true", methodName)
		return
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Errorf("%s() error is not a status error", methodName)
		return
	}
	if st.Code() != wantCode {
		t.Errorf("%s() code = %v, want %v", methodName, st.Code(), wantCode)
	}
}

// Helper function to verify no error occurred.
func checkNoError(t *testing.T, err error, methodName string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s() unexpected error = %v", methodName, err)
	}
}

// Helper function to verify response is not nil.
func checkResponseNotNil(t *testing.T, resp interface{}, methodName string) bool {
	t.Helper()
	if resp == nil {
		t.Errorf("%s() response is nil", methodName)
		return false
	}
	return true
}

// Helper function to compare profile response fields.
func compareProfileFields(t *testing.T, got, want *betterreads.GetCurrentUserProfileResponse, methodName string) {
	t.Helper()
	if got.Id != want.Id {
		t.Errorf("%s() Id = %v, want %v", methodName, got.Id, want.Id)
	}
	if got.Username != want.Username {
		t.Errorf("%s() Username = %v, want %v", methodName, got.Username, want.Username)
	}
	if got.Email != want.Email {
		t.Errorf("%s() Email = %v, want %v", methodName, got.Email, want.Email)
	}
	if got.FirstName != want.FirstName {
		t.Errorf("%s() FirstName = %v, want %v", methodName, got.FirstName, want.FirstName)
	}
	if got.LastName != want.LastName {
		t.Errorf("%s() LastName = %v, want %v", methodName, got.LastName, want.LastName)
	}
	if got.ProfilePhotoUrl != want.ProfilePhotoUrl {
		t.Errorf("%s() ProfilePhotoUrl = %v, want %v", methodName, got.ProfilePhotoUrl, want.ProfilePhotoUrl)
	}
}

// Helper function to compare update profile response fields.
func compareUpdateProfileFields(t *testing.T, got, want *betterreads.UpdateUserProfileResponse, methodName string) {
	t.Helper()
	if got.Id != want.Id {
		t.Errorf("%s() Id = %v, want %v", methodName, got.Id, want.Id)
	}
	if got.Username != want.Username {
		t.Errorf("%s() Username = %v, want %v", methodName, got.Username, want.Username)
	}
	if got.Email != want.Email {
		t.Errorf("%s() Email = %v, want %v", methodName, got.Email, want.Email)
	}
}

// Helper function to compare create profile response fields.
func compareCreateProfileFields(t *testing.T, got, want *betterreads.CreateUserProfileResponse, methodName string) {
	t.Helper()
	if got.Id != want.Id {
		t.Errorf("%s() Id = %v, want %v", methodName, got.Id, want.Id)
	}
	if got.Username != want.Username {
		t.Errorf("%s() Username = %v, want %v", methodName, got.Username, want.Username)
	}
	if got.Email != want.Email {
		t.Errorf("%s() Email = %v, want %v", methodName, got.Email, want.Email)
	}
}

func TestServer_DeleteUserProfile(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	tests := []struct {
		name      string
		ctx       context.Context
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "successful deletion",
			ctx:  ctx,
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileDelete(gomock.Any(), testUserID).
					Return(nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name:      "missing user_id in context",
			ctx:       ctxNoUserID,
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "user not found",
			ctx:  ctx,
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileDelete(gomock.Any(), testUserID).
					Return(postgres.ErrUserNotFound)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name: "database error",
			ctx:  ctx,
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileDelete(gomock.Any(), testUserID).
					Return(errors.New("database connection failed"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockAPI(ctrl)
			tt.setupMock(mockDB)

			s := &Server{DB: mockDB}

			resp, err := s.DeleteUserProfile(tt.ctx, &betterreads.DeleteUserProfileRequest{})

			if tt.wantErr {
				checkErrorCode(t, err, tt.wantCode, "DeleteUserProfile")
				return
			}

			checkNoError(t, err, "DeleteUserProfile")
			checkResponseNotNil(t, resp, "DeleteUserProfile")
		})
	}
}

func TestServer_GetCurrentUserProfile(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	testProfile := &data.User{
		ID:              testUserID,
		Username:        "testuser",
		FirstName:       "Test",
		LastName:        "User",
		Email:           "test@example.com",
		ProfilePhotoURL: "https://example.com/photo.jpg",
		CreatedAt:       testTime,
	}

	tests := []struct {
		name      string
		ctx       context.Context
		setupMock func(*mocks.MockAPI)
		wantResp  *betterreads.GetCurrentUserProfileResponse
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "successful retrieval",
			ctx:  ctx,
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileGet(gomock.Any(), testUserID).
					Return(testProfile, nil)
			},
			wantResp: &betterreads.GetCurrentUserProfileResponse{
				CreatedAt:       timestamppb.New(testTime),
				Email:           "test@example.com",
				FirstName:       "Test",
				Id:              testUserID,
				LastName:        "User",
				ProfilePhotoUrl: "https://example.com/photo.jpg",
				Username:        "testuser",
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name:      "missing user_id in context",
			ctx:       ctxNoUserID,
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "user not found",
			ctx:  ctx,
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileGet(gomock.Any(), testUserID).
					Return(nil, postgres.ErrUserNotFound)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name: "database error",
			ctx:  ctx,
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileGet(gomock.Any(), testUserID).
					Return(nil, errors.New("database connection failed"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockAPI(ctrl)
			tt.setupMock(mockDB)

			s := &Server{DB: mockDB}

			resp, err := s.GetCurrentUserProfile(tt.ctx, &betterreads.GetCurrentUserProfileRequest{})

			if tt.wantErr {
				checkErrorCode(t, err, tt.wantCode, "GetCurrentUserProfile")
				return
			}

			checkNoError(t, err, "GetCurrentUserProfile")
			if !checkResponseNotNil(t, resp, "GetCurrentUserProfile") {
				return
			}

			compareProfileFields(t, resp, tt.wantResp, "GetCurrentUserProfile")
		})
	}
}

func TestServer_UpdateUserProfile(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	updatedProfile := &data.User{
		ID:              testUserID,
		Username:        "updateduser",
		FirstName:       "Updated",
		LastName:        "User",
		Email:           "updated@example.com",
		ProfilePhotoURL: "https://example.com/new-photo.jpg",
		CreatedAt:       testTime,
	}

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.UpdateUserProfileRequest
		setupMock func(*mocks.MockAPI)
		wantResp  *betterreads.UpdateUserProfileResponse
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "successful update",
			ctx:  ctx,
			request: &betterreads.UpdateUserProfileRequest{
				Username:        "updateduser",
				FirstName:       "Updated",
				LastName:        "User",
				Email:           "updated@example.com",
				ProfilePhotoUrl: "https://example.com/new-photo.jpg",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileUpdate(gomock.Any(), testUserID, gomock.Any()).
					Return(updatedProfile, nil)
			},
			wantResp: &betterreads.UpdateUserProfileResponse{
				CreatedAt:       timestamppb.New(testTime),
				Email:           "updated@example.com",
				FirstName:       "Updated",
				Id:              testUserID,
				LastName:        "User",
				ProfilePhotoUrl: "https://example.com/new-photo.jpg",
				Username:        "updateduser",
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.UpdateUserProfileRequest{
				Username: "updateduser",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "user not found",
			ctx:  ctx,
			request: &betterreads.UpdateUserProfileRequest{
				Username: "updateduser",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileUpdate(gomock.Any(), testUserID, gomock.Any()).
					Return(nil, postgres.ErrUserNotFound)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name: "username already exists",
			ctx:  ctx,
			request: &betterreads.UpdateUserProfileRequest{
				Username: "existinguser",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileUpdate(gomock.Any(), testUserID, gomock.Any()).
					Return(nil, postgres.ErrUserNameExists)
			},
			wantCode: codes.AlreadyExists,
			wantErr:  true,
		},
		{
			name: "email already exists",
			ctx:  ctx,
			request: &betterreads.UpdateUserProfileRequest{
				Email: "existing@example.com",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileUpdate(gomock.Any(), testUserID, gomock.Any()).
					Return(nil, postgres.ErrEmailExists)
			},
			wantCode: codes.AlreadyExists,
			wantErr:  true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.UpdateUserProfileRequest{
				Username: "updateduser",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileUpdate(gomock.Any(), testUserID, gomock.Any()).
					Return(nil, errors.New("database connection failed"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockAPI(ctrl)
			tt.setupMock(mockDB)

			s := &Server{DB: mockDB}

			resp, err := s.UpdateUserProfile(tt.ctx, tt.request)

			if tt.wantErr {
				checkErrorCode(t, err, tt.wantCode, "UpdateUserProfile")
				return
			}

			checkNoError(t, err, "UpdateUserProfile")
			if !checkResponseNotNil(t, resp, "UpdateUserProfile") {
				return
			}

			compareUpdateProfileFields(t, resp, tt.wantResp, "UpdateUserProfile")
		})
	}
}

func TestServer_CreateUserProfile(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	createdProfile := &data.User{
		ID:              testUserID,
		Username:        "newuser",
		FirstName:       "New",
		LastName:        "User",
		Email:           "new@example.com",
		ProfilePhotoURL: "https://example.com/photo.jpg",
		CreatedAt:       testTime,
	}

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.CreateUserProfileRequest
		setupMock func(*mocks.MockAPI)
		wantResp  *betterreads.CreateUserProfileResponse
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "successful creation",
			ctx:  ctx,
			request: &betterreads.CreateUserProfileRequest{
				Username:        "newuser",
				FirstName:       "New",
				LastName:        "User",
				Email:           "new@example.com",
				ProfilePhotoUrl: "https://example.com/photo.jpg",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileCreate(gomock.Any(), gomock.Any()).
					Return(createdProfile, nil)
			},
			wantResp: &betterreads.CreateUserProfileResponse{
				CreatedAt:       timestamppb.New(testTime),
				Email:           "new@example.com",
				FirstName:       "New",
				Id:              testUserID,
				LastName:        "User",
				ProfilePhotoUrl: "https://example.com/photo.jpg",
				Username:        "newuser",
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.CreateUserProfileRequest{
				Username:  "newuser",
				FirstName: "New",
				LastName:  "User",
				Email:     "new@example.com",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "username too short",
			ctx:  ctx,
			request: &betterreads.CreateUserProfileRequest{
				Username:  "ab",
				FirstName: "New",
				LastName:  "User",
				Email:     "new@example.com",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "invalid email",
			ctx:  ctx,
			request: &betterreads.CreateUserProfileRequest{
				Username:  "newuser",
				FirstName: "New",
				LastName:  "User",
				Email:     "invalid-email",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "username already exists",
			ctx:  ctx,
			request: &betterreads.CreateUserProfileRequest{
				Username:  "existinguser",
				FirstName: "New",
				LastName:  "User",
				Email:     "new@example.com",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileCreate(gomock.Any(), gomock.Any()).
					Return(nil, postgres.ErrUserNameExists)
			},
			wantCode: codes.AlreadyExists,
			wantErr:  true,
		},
		{
			name: "email already exists",
			ctx:  ctx,
			request: &betterreads.CreateUserProfileRequest{
				Username:  "newuser",
				FirstName: "New",
				LastName:  "User",
				Email:     "existing@example.com",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileCreate(gomock.Any(), gomock.Any()).
					Return(nil, postgres.ErrEmailExists)
			},
			wantCode: codes.AlreadyExists,
			wantErr:  true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.CreateUserProfileRequest{
				Username:  "newuser",
				FirstName: "New",
				LastName:  "User",
				Email:     "new@example.com",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					ProfileCreate(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database connection failed"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockAPI(ctrl)
			tt.setupMock(mockDB)

			s := &Server{DB: mockDB}

			resp, err := s.CreateUserProfile(tt.ctx, tt.request)

			if tt.wantErr {
				checkErrorCode(t, err, tt.wantCode, "CreateUserProfile")
				return
			}

			checkNoError(t, err, "CreateUserProfile")
			if !checkResponseNotNil(t, resp, "CreateUserProfile") {
				return
			}

			compareCreateProfileFields(t, resp, tt.wantResp, "CreateUserProfile")
		})
	}
}

func Test_isValidEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "valid email",
			email: "test@example.com",
			want:  true,
		},
		{
			name:  "valid email with subdomain",
			email: "user@mail.example.com",
			want:  true,
		},
		{
			name:  "valid email with plus",
			email: "user+tag@example.com",
			want:  true,
		},
		{
			name:  "invalid email - no @",
			email: "invalid-email",
			want:  false,
		},
		{
			name:  "invalid email - no domain",
			email: "user@",
			want:  false,
		},
		{
			name:  "invalid email - no user",
			email: "@example.com",
			want:  false,
		},
		{
			name:  "invalid email - spaces",
			email: "user name@example.com",
			want:  false,
		},
		{
			name:  "empty email",
			email: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isValidEmail(tt.email)
			if got != tt.want {
				t.Errorf("isValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}
