package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	MaxMessageContentLength = 5000
)

// room member roles
type RoomMemberRole string

const (
	AdminRole RoomMemberRole = "admin"
	Member    RoomMemberRole = "member"
)

type Room struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatorID uuid.UUID `json:"creatorId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateRoomReq struct {
	Name   string    `json:"name"`
	UserID uuid.UUID `json:"userId"`
}

func (r *CreateRoomReq) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("room name is required")
	}
	if r.UserID == uuid.Nil {
		return fmt.Errorf("user id is required")
	}
	return nil
}

type RoomMember struct {
	ID        uuid.UUID `json:"id"`
	RoomID    uuid.UUID `json:"roomId"`
	UserID    uuid.UUID `json:"userId"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateRoomMemberReq struct {
	UserID uuid.UUID `json:"userId"`
}

func (r *CreateRoomMemberReq) Validate() error {
	if r.UserID == uuid.Nil {
		return fmt.Errorf("user id is required")
	}
	return nil
}

type Message struct {
	ID             uuid.UUID `json:"id"`
	RoomID         uuid.UUID `json:"roomId"`
	SenderID       uuid.UUID `json:"senderId"`
	SenderUsername string    `json:"senderUsername"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type CreateMessageReq struct {
	Content string `json:"content"`
}

func (m *CreateMessageReq) Validate() error {
	if m.Content == "" {
		return fmt.Errorf("room name is required")
	}
	if len(m.Content) > MaxMessageContentLength {
		return fmt.Errorf("content has exceeded its limit of %v characters", MaxMessageContentLength)
	}
	return nil
}
