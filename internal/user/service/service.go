package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/internal/user/model"
	"github.com/CimaCha/gophermart-loyal-service/internal/user/storage"
	"github.com/google/uuid"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

//go:generate go tool mockgen -source=service.go -destination=mock/dependencies_gen.go -package=mock

type TokenBuilder interface {
	BuildJWTString(uuid.UUID) (string, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password, hash string) (bool, error)
}

type UserService struct {
	storage        storage.UserStorage
	tokenBuilder   TokenBuilder
	passwordHasher PasswordHasher
}

func NewUserService(userStorage storage.UserStorage, tokenBuilder TokenBuilder, passwordHasher PasswordHasher) *UserService {
	return &UserService{
		storage:        userStorage,
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
	err = s.storage.SaveUserInfo(ctx, model.UserInfo{
		UUID:         userID,
		Login:        userLogin,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			return "", ErrUserAlreadyExists
		}
		return "", fmt.Errorf("save user: %w", err)
	}

	return token, nil
}

func (s *UserService) LoginUser(ctx context.Context, userLogin string, password string) (string, error) {
	userInfo, err := s.storage.FindUserInfo(ctx, userLogin)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
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
