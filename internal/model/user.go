package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	MaxUsernameLength = 50
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateUserReq struct {
	Username string `json:"username"`
}

func (r *CreateUserReq) Validate() error {
	if r.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(r.Username) > MaxUsernameLength {
		return fmt.Errorf("username cannot exceed %d characters", MaxUsernameLength)
	}
	return nil
}
