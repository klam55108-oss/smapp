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

type User struct {
	db *sql.DB
}

func NewUser(db *sql.DB) *User {
	return &User{db: db}
}

var (
	ErrEmailExists  = errors.New("email already exists")
	ErrHandleExists = errors.New("handle already exists")
)

func (u *User) Create(ctx context.Context, name, email, handle, passwordHash string) (uuid.UUID, error) {
	fail := func(err error) (uuid.UUID, error) {
		return uuid.Nil, fmt.Errorf("insert user in db: %w", err)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return fail(err)
	}
	_, err = u.db.ExecContext(
		ctx,
		"INSERT INTO users (id, name, email, handle, password_hash) VALUES (?, ?, ?, ?, ?)",
		id[:], name, email, handle, passwordHash,
	)
	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
		if strings.Contains(mysqlError.Message, "email_unique") {
			return uuid.Nil, ErrEmailExists
		} else if strings.Contains(mysqlError.Message, "handle_unique") {
			return uuid.Nil, ErrHandleExists
		} else {
			return fail(fmt.Errorf("unexpected unique constraint: %w", err))
		}
	}
	if err != nil {
		return fail(err)
	}

	return id, nil
}

func (u *User) GetAuthData(ctx context.Context, identifier string) (uuid.UUID, []byte, error) {
	fail := func(err error) (uuid.UUID, []byte, error) {
		return uuid.Nil, nil, fmt.Errorf("get auth data from db: %w", err)
	}

	var id uuid.UUID
	var passwordHash []byte
	err := u.db.QueryRowContext(
		ctx,
		fmt.Sprintf("SELECT id, password_hash FROM users WHERE %s = ?", getIdentiferType(identifier)),
		identifier,
	).Scan(&id, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, nil, ErrRecordNotFound
	}
	if err != nil {
		return fail(err)
	}
	return id, passwordHash, nil
}

func getIdentiferType(identifier string) string {
	if strings.ContainsRune(identifier, '@') {
		return "email"
	}
	return "handle"
}
