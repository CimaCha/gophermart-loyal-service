package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/internal/user/model"
	userrepo "github.com/CimaCha/gophermart-loyal-service/internal/user/repository"
	"github.com/google/uuid"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type TokenBuilder interface {
	BuildJWTString(uuid.UUID) (string, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password, hash string) (bool, error)
}

type UserRepository interface {
	FindUserInfo(ctx context.Context, userLogin string) (*model.UserInfo, error)
	SaveUserInfo(ctx context.Context, userInfo model.UserInfo) error
}

type UserService struct {
	repo           UserRepository
	tokenBuilder   TokenBuilder
	passwordHasher PasswordHasher
}

func New(userRepo UserRepository, tokenBuilder TokenBuilder, passwordHasher PasswordHasher) *UserService {
	return &UserService{
		repo:           userRepo,
		tokenBuilder:   tokenBuilder,
		passwordHasher: passwordHasher,
	}
}

func (s *UserService) CreateUser(ctx context.Context, userLogin string, password string) (string, error) {
	userID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	passwordHash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	token, err := s.tokenBuilder.BuildJWTString(userID)
	if err != nil {
		return "", fmt.Errorf("build token: %w", err)
	}
	err = s.repo.SaveUserInfo(ctx, model.UserInfo{
		UUID:         userID,
		Login:        userLogin,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		if errors.Is(err, userrepo.ErrUserAlreadyExists) {
			return "", ErrUserAlreadyExists
		}
		return "", fmt.Errorf("save user: %w", err)
	}

	return token, nil
}

func (s *UserService) LoginUser(ctx context.Context, userLogin string, password string) (string, error) {
	userInfo, err := s.repo.FindUserInfo(ctx, userLogin)
	if err != nil {
		if errors.Is(err, userrepo.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("find user: %w", err)
	}

	match, err := s.passwordHasher.Compare(password, userInfo.PasswordHash)
	if err != nil {
		return "", fmt.Errorf("compare password: %w", err)
	}
	if !match {
		return "", ErrInvalidCredentials
	}

	token, err := s.tokenBuilder.BuildJWTString(userInfo.UUID)
	if err != nil {
		return "", fmt.Errorf("build token: %w", err)
	}
	return token, nil
}
