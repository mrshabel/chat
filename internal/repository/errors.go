package repository

import "errors"

// repository specific errors
var (
	ErrNotFound     = errors.New("not found")
	ErrAlreadyExist = errors.New("already exists")
)
