package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var (
	ErrEmailExists  = errors.New("email already exists")
	ErrHandleExists = errors.New("handle already exists")
	ErrNoSuchUser   = errors.New("no such user")
)

type User struct {
	db *sql.DB
}

func (u *User) Close() error {
	return u.db.Close()
}

func NewUser(db *sql.DB) *User {
	return &User{db: db}
}

func (u *User) Create(ctx context.Context, name, email, handle, password_hash string) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("insert user in db: %w", err)
	}
	_, err = u.db.ExecContext(
		ctx,
		"INSERT INTO users (id, name, email, handle, password_hash) VALUES (?, ?, ?, ?, ?)",
		id[:], name, email, handle, password_hash,
	)
	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
		if strings.Contains(mysqlError.Message, "email_unique") {
			return "", fmt.Errorf("insert user in db: %w", ErrEmailExists)
		} else if strings.Contains(mysqlError.Message, "handle_unique") {
			return "", fmt.Errorf("insert user in db: %w", ErrHandleExists)
		} else {
			return "", fmt.Errorf("insert user in db: unexprected unique constraint: %w", err)
		}
	}
	if err != nil {
		return "", fmt.Errorf("insert user in db: %w", err)
	}

	return id.String(), nil
}

func (u *User) GetAuthData(ctx context.Context, identifier string) (string, []byte, error) {
	var id uuid.UUID
	var passwordHash []byte
	err := u.db.QueryRowContext(
		ctx,
		fmt.Sprintf("SELECT id, password_hash FROM users WHERE %s = ?", getIdentiferType(identifier)),
		identifier,
	).Scan(&id, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil, fmt.Errorf("get auth data from db: %w", ErrNoSuchUser)
	}
	if err != nil {
		return "", nil, fmt.Errorf("get auth data from db: %w", err)
	}
	return id.String(), passwordHash, nil
}

func getIdentiferType(identifier string) string {
	if strings.ContainsRune(identifier, '@') {
		return "email"
	}
	return "handle"
}
