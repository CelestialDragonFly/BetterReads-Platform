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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_GetUserById(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()

	testUser := &data.User{
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
		userID    string
		setupMock func(*mocks.MockAPI)
		wantResp  *betterreads.GetUserByIdResponse
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name:   "successful retrieval",
			userID: testUserID,
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserByID(gomock.Any(), testUserID).
					Return(testUser, nil)
			},
			wantResp: &betterreads.GetUserByIdResponse{
				Id:              testUserID,
				Username:        "testuser",
				FirstName:       "Test",
				LastName:        "User",
				ProfilePhotoUrl: "https://example.com/photo.jpg",
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name:   "user not found",
			userID: "non-existent-user",
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserByID(gomock.Any(), "non-existent-user").
					Return(nil, postgres.ErrUserNotFound)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name:   "database error",
			userID: testUserID,
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserByID(gomock.Any(), testUserID).
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
			resp, err := s.GetUserById(ctx, &betterreads.GetUserByIdRequest{UserId: tt.userID})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok, "error should be a status error")
				assert.Equal(t, tt.wantCode, st.Code())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.wantResp.Id, resp.Id)
			assert.Equal(t, tt.wantResp.Username, resp.Username)
			assert.Equal(t, tt.wantResp.FirstName, resp.FirstName)
			assert.Equal(t, tt.wantResp.LastName, resp.LastName)
			assert.Equal(t, tt.wantResp.ProfilePhotoUrl, resp.ProfilePhotoUrl)
		})
	}
}

func TestServer_FollowUser(t *testing.T) {
	t.Parallel()

	followerID := "follower-123"
	followeeID := "followee-456"
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, followerID)
	ctxNoUserID := context.Background()

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.FollowUserRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "successful follow",
			ctx:  ctx,
			request: &betterreads.FollowUserRequest{
				UserId: followeeID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					FollowUser(gomock.Any(), followerID, followeeID).
					Return(nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.FollowUserRequest{
				UserId: followeeID,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "self follow",
			ctx:  ctx,
			request: &betterreads.FollowUserRequest{
				UserId: followerID, // Trying to follow self (should pass correct ID for mock expectations if checking actual call params, or mock checks it)
				// Wait, the Server impl passes (ctx, "follower-123", "follower-123") to DB if I pass same ID.
				// But typically the DB or service logic would raise ErrSelfFollow.
				// Here I am testing the Server logic that handles the error returned by DB.
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					FollowUser(gomock.Any(), followerID, followerID).
					Return(postgres.ErrSelfFollow)
			},
			wantCode: codes.InvalidArgument,
			wantErr:  true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.FollowUserRequest{
				UserId: followeeID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					FollowUser(gomock.Any(), followerID, followeeID).
					Return(errors.New("database error"))
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
			resp, err := s.FollowUser(tt.ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok, "error should be a status error")
				assert.Equal(t, tt.wantCode, st.Code())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func TestServer_UnfollowUser(t *testing.T) {
	t.Parallel()

	followerID := "follower-123"
	followeeID := "followee-456"
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, followerID)
	ctxNoUserID := context.Background()

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.UnfollowUserRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "successful unfollow",
			ctx:  ctx,
			request: &betterreads.UnfollowUserRequest{
				UserId: followeeID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					UnfollowUser(gomock.Any(), followerID, followeeID).
					Return(nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.UnfollowUserRequest{
				UserId: followeeID,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "self unfollow",
			ctx:  ctx,
			request: &betterreads.UnfollowUserRequest{
				UserId: followerID, // Trying to unfollow self
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					UnfollowUser(gomock.Any(), followerID, followerID).
					Return(postgres.ErrSelfFollow)
			},
			wantCode: codes.InvalidArgument,
			wantErr:  true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.UnfollowUserRequest{
				UserId: followeeID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					UnfollowUser(gomock.Any(), followerID, followeeID).
					Return(errors.New("database error"))
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
			resp, err := s.UnfollowUser(tt.ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok, "error should be a status error")
				assert.Equal(t, tt.wantCode, st.Code())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}
