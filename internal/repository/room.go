package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/mrshabel/chat/internal/model"
)

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(ctx context.Context, data *model.Room) (*model.Room, error) {
	query := `
        INSERT INTO rooms (name, creator_id)
        VALUES ($1, $2)
		RETURNING id, name, creator_id, created_at, updated_at
    `
	var room model.Room
	if err := r.db.QueryRowContext(ctx, query, data.Name, data.CreatorID).Scan(&room.ID, &room.Name, &room.CreatorID, &room.CreatedAt, &room.UpdatedAt); err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Room, error) {
	query := `
        SELECT id, name, creator_id, created_at, updated_at 
        FROM rooms 
        WHERE id = $1
    `
	var room model.Room
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&room.ID,
		&room.Name,
		&room.CreatorID,
		&room.CreatedAt,
		&room.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return &room, err
}

func (r *RoomRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.Room, error) {
	query := `
        SELECT id, name, creator_id, created_at, updated_at 
        FROM rooms 
        ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
    `
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*model.Room
	for rows.Next() {
		var room model.Room
		if err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.CreatorID,
			&room.CreatedAt,
			&room.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *RoomRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Room, error) {
	query := `
        SELECT r.id, r.name, r.creator_id, r.created_at, r.updated_at 
        FROM rooms r
		LEFT JOIN room_members rm
		ON rm.room_id = r.id
		WHERE rm.user_id = $1
        ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*model.Room
	for rows.Next() {
		var room model.Room
		if err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.CreatorID,
			&room.CreatedAt,
			&room.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rooms = append(rooms, &room)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *RoomRepository) AddMember(ctx context.Context, roomID, userID uuid.UUID, role string) (*model.RoomMember, error) {
	query := `
        INSERT INTO room_members (room_id, user_id, role)
        VALUES ($1, $2, $3)
		RETURNING id, room_id, user_id, role, created_at, updated_at
    `
	var member model.RoomMember
	if err := r.db.QueryRowContext(ctx, query, roomID, userID, role).Scan(&member.ID, &member.RoomID, &member.UserID, &member.Role, &member.CreatedAt, &member.UpdatedAt); err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *RoomRepository) GetAllMembers(ctx context.Context, roomID string, limit, offset int) ([]*model.RoomMember, error) {
	query := `
        SELECT id, room_id, user_id, role, created_at, updated_at
        FROM room_members
        WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.RoomMember
	for rows.Next() {
		var member model.RoomMember
		if err := rows.Scan(
			&member.ID,
			&member.RoomID,
			&member.UserID,
			&member.Role,
			&member.CreatedAt,
			&member.UpdatedAt,
		); err != nil {
			return nil, err
		}
		members = append(members, &member)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}
