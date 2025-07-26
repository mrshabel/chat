package model

import (
	"fmt"
	"time"
)

const (
	MaxMessageContentLength = 5000
)

type Room struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateRoomReq struct {
	Name string `json:"name"`
}

func (r *CreateRoomReq) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("room name is required")
	}
	return nil
}

type Message struct {
	ID              string    `json:"id"`
	RoomID          string    `json:"roomId"`
	CreatorID       string    `json:"creatorId"`
	CreatorUsername string    `json:"creatorUsername"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"createdAt"`
}

type CreateMessageReq struct {
	ID              string    `json:"id"`
	RoomID          string    `json:"roomId"`
	CreatorID       string    `json:"creatorId"`
	CreatorUsername string    `json:"creatorUsername"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"createdAt"`
}

func (r *CreateMessageReq) Validate() error {
	if r.RoomID == "" {
		return fmt.Errorf("room ID is required")
	}
	if r.Content == "" {
		return fmt.Errorf("content is required")
	}
	if len(r.Content) > MaxMessageContentLength {
		return fmt.Errorf("content has exceeded its limit of %v characters", MaxMessageContentLength)
	}
	return nil
}

type Client struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
