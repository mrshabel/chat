package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/mrshabel/chat/internal/model"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, data *model.Message) (*model.Message, error) {
	query := `
        INSERT INTO messages (room_id, sender_id, sender_username, content)
        VALUES ($1, $2, $3, $4)
		RETURNING id, room_id, sender_id, sender_username, content, created_at, updated_at
    `
	var message model.Message
	if err := r.db.QueryRowContext(ctx, query, data.RoomID, data.SenderID, data.SenderUsername, data.Content).Scan(
		&message.ID,
		&message.RoomID,
		&message.SenderID,
		&message.SenderUsername,
		&message.Content,
		&message.CreatedAt,
		&message.UpdatedAt); err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *MessageRepository) GetByRoomID(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*model.Message, error) {
	query := `
        SELECT id, room_id, sender_id, sender_username, content, created_at, updated_at
        FROM messages 
        WHERE room_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.SenderID,
			&msg.SenderUsername,
			&msg.Content,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	if err = rows.Err(); err != nil {
		return messages, nil
	}
	return messages, nil
}

func (r *MessageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM messages WHERE id = $1"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
