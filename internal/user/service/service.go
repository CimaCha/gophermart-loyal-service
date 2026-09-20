package service

import (
	"context"
	"errors"
	"github.com/CimaCha/gophermart-loyal-service/internal/user/model"
	storage2 "github.com/CimaCha/gophermart-loyal-service/internal/user/storage"
	"github.com/google/uuid"
	"time"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid password")
)

type TokenBuilder interface {
	BuildJWTString(uuid.UUID) (string, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password, hash string) (bool, error)
}
type UserService struct {
	storage        storage2.UserStorage
	tokenBuilder   TokenBuilder
	passwordHasher PasswordHasher
}

func NewUserService(storage storage2.UserStorage, tokenBuilder TokenBuilder, passwordHasher PasswordHasher) UserService {
	return UserService{
		storage:        storage,
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
		return "", err
	}

	token, err := s.tokenBuilder.BuildJWTString(userID)
	if err != nil {
		return "", err
	}
	err = s.storage.SaveUserInfo(ctx, model.UserInfo{
		UUID:         userID,
		Login:        userLogin,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) LoginUser(ctx context.Context, userLogin string, password string) (string, error) {
	userInfo, err := s.storage.FindUserInfo(ctx, userLogin)
	if err != nil {
		return "", err
	}

	match, err := s.passwordHasher.Compare(password, userInfo.PasswordHash)
	if err != nil {
		return "", err
	}
	if !match {
		return "", ErrInvalidCredentials
	}

	return s.tokenBuilder.BuildJWTString(userInfo.UUID)
}
