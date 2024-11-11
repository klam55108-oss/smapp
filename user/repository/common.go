package repository

import "errors"

var (
	ErrUserIDNotFound = errors.New("id not found in users table")
	ErrRecordNotFound = errors.New("record not found")
	ErrRecordExists   = errors.New("record already exists")
)
