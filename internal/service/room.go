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
	ErrRoomAlreadyExist = errors.New("room already exist")
	ErrRoomNotFound     = errors.New("room not found")
)

type RoomService struct {
	repo *repository.RoomRepository
}

func NewRoomService(repo *repository.RoomRepository) *RoomService {
	return &RoomService{repo: repo}
}

func (s *RoomService) Create(ctx context.Context, req *model.CreateRoomReq) (*model.Room, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	room := &model.Room{
		Name:      req.Name,
		CreatorID: req.UserID,
	}

	room, err := s.repo.Create(ctx, room)
	if err != nil {
		return nil, err
	}

	// add creator as admin
	if _, err := s.repo.AddMember(ctx, room.ID, room.CreatorID, string(model.AdminRole)); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *RoomService) GetByID(ctx context.Context, id uuid.UUID) (*model.Room, error) {
	room, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			err = ErrRoomNotFound
		}
		return nil, err
	}
	return room, nil
}

func (s *RoomService) GetAllByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Room, error) {
	return s.repo.GetAllByUserID(ctx, userID, limit, offset)
}

func (s *RoomService) GetAll(ctx context.Context, limit, offset int) ([]*model.Room, error) {
	return s.repo.GetAll(ctx, limit, offset)
}

func (s *RoomService) AddMember(ctx context.Context, roomID, userID uuid.UUID, role string) (*model.RoomMember, error) {
	return s.repo.AddMember(ctx, roomID, userID, role)
}

func (s *RoomService) GetAllMembers(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*model.RoomMember, error) {
	return s.repo.GetAllMembers(ctx, roomID.String(), limit, offset)
}
