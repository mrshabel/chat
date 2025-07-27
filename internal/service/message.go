package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mrshabel/chat/internal/model"
	"github.com/mrshabel/chat/internal/repository"
)

type MessageService struct {
	repo *repository.MessageRepository
}

func NewMessageService(repo *repository.MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

func (s *MessageService) Create(ctx context.Context, msg *model.Message) (*model.Message, error) {
	return s.repo.Create(ctx, msg)
}

// GetByRoomID retrieves all messages for a given room
func (s *MessageService) GetByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*model.Message, error) {
	return s.repo.GetByRoomID(ctx, roomID, limit, offset)
}

func (s *MessageService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
