package service

import (
	"context"
	"errors"
	"testing"

	"github.com/CimaCha/gophermart-loyal-service/internal/user/model"
	userrepo "github.com/CimaCha/gophermart-loyal-service/internal/user/repository"
	"github.com/CimaCha/gophermart-loyal-service/internal/user/service/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testStruct struct{}

func TestUserService_CreateUser(t *testing.T) {
	errHash := errors.New("hash failed")
	errToken := errors.New("token failed")
	errRepository := errors.New("storage failed")

	tests := []struct {
		name      string
		hashErr   error
		tokenErr  error
		repoErr   error
		wantErr   error
		wantToken string
	}{
		{name: "hash error", hashErr: errHash, wantErr: errHash},
		{name: "token error", tokenErr: errToken, wantErr: errToken},
		{name: "duplicate user", repoErr: userrepo.ErrUserAlreadyExists, wantErr: ErrUserAlreadyExists},
		{name: "storage error", repoErr: errRepository, wantErr: errRepository},
		{name: "success", wantToken: "signed-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), testStruct{}, "request")
			ctrl := gomock.NewController(t)
			storageMock := mock.NewMockUserStorage(ctrl)
			tokenBuilderMock := mock.NewMockTokenBuilder(ctrl)
			passwordHasherMock := mock.NewMockPasswordHasher(ctrl)

			calls := []any{
				passwordHasherMock.EXPECT().Hash("secret").Return("encoded-hash", tt.hashErr),
			}
			if tt.hashErr == nil {
				calls = append(calls, tokenBuilderMock.EXPECT().
					BuildJWTString(gomock.Any()).
					DoAndReturn(func(userID uuid.UUID) (string, error) {
						require.NotEqual(t, uuid.Nil, userID)
						return "signed-token", tt.tokenErr
					}))
			}
			if tt.hashErr == nil && tt.tokenErr == nil {
				calls = append(calls, storageMock.EXPECT().
					SaveUserInfo(gomock.Any(), gomock.Any()).
					DoAndReturn(func(gotCtx context.Context, user model.UserInfo) error {
						require.Same(t, ctx, gotCtx)
						require.NotEqual(t, uuid.Nil, user.UUID)
						require.Equal(t, "alice", user.Login)
						require.Equal(t, "encoded-hash", user.PasswordHash)
						require.NotEqual(t, "secret", user.PasswordHash)
						require.False(t, user.CreatedAt.IsZero())
						return tt.repoErr
					}))
			}
			gomock.InOrder(calls...)

			svc := New(
				storageMock,
				tokenBuilderMock,
				passwordHasherMock,
			)

			got, err := svc.CreateUser(ctx, "alice", "secret")

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantToken, got)
			}
		})
	}
}

func TestUserService_LoginUser(t *testing.T) {
	errFind := errors.New("find failed")
	errCompare := errors.New("compare failed")
	errToken := errors.New("token failed")
	userID := uuid.New()

	tests := []struct {
		name       string
		findErr    error
		compareErr error
		match      bool
		tokenErr   error
		wantErr    error
		wantToken  string
	}{
		{name: "user not found", findErr: userrepo.ErrUserNotFound, wantErr: ErrInvalidCredentials},
		{name: "storage error", findErr: errFind, wantErr: errFind},
		{name: "compare error", compareErr: errCompare, wantErr: errCompare},
		{name: "password mismatch", wantErr: ErrInvalidCredentials},
		{name: "token error", match: true, tokenErr: errToken, wantErr: errToken},
		{name: "success", match: true, wantToken: "signed-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), testStruct{}, "request")
			ctrl := gomock.NewController(t)
			storageMock := mock.NewMockUserStorage(ctrl)
			tokenBuilderMock := mock.NewMockTokenBuilder(ctrl)
			passwordHasherMock := mock.NewMockPasswordHasher(ctrl)

			calls := []any{
				storageMock.EXPECT().
					FindUserInfo(gomock.Any(), "alice").
					DoAndReturn(func(gotCtx context.Context, login string) (model.UserInfo, error) {
						require.Same(t, ctx, gotCtx)
						return model.UserInfo{UUID: userID, Login: login, PasswordHash: "encoded-hash"}, tt.findErr
					}),
			}
			if tt.findErr == nil {
				calls = append(calls, passwordHasherMock.EXPECT().
					Compare("secret", "encoded-hash").
					Return(tt.match, tt.compareErr))
			}
			if tt.findErr == nil && tt.compareErr == nil && tt.match {
				calls = append(calls, tokenBuilderMock.EXPECT().
					BuildJWTString(userID).
					Return("signed-token", tt.tokenErr))
			}
			gomock.InOrder(calls...)

			svc := New(
				storageMock,
				tokenBuilderMock,
				passwordHasherMock,
			)

			got, err := svc.LoginUser(ctx, "alice", "secret")

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantToken, got)
			}
		})
	}
}
