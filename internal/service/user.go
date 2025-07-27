package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mrshabel/chat/internal/model"
	"github.com/mrshabel/chat/internal/repository"
)

// errors
var (
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrUserNotFound     = errors.New("user not found")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, req *model.CreateUserReq) (*model.User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	// Check if username exists
	existing, err := s.repo.GetByUsername(ctx, req.Username)
	if existing != nil && err == nil {
		return nil, err
	}

	return s.repo.Create(ctx, req.Username)
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			err = ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			err = ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
