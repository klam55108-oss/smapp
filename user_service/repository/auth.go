package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var (
	ErrEmailExists     = errors.New("email already exists")
	ErrHandleExists    = errors.New("handle already exists")
	ErrNoSuchUser      = errors.New("no such user")
	ErrDatabaseTimeout = errors.New("database timeout")
)

type User struct {
	db                 *sql.DB
	insertQueryTimeout time.Duration
	getQueryTimeout    time.Duration
}

func (u *User) Close() error {
	return u.db.Close()
}

func NewUser(db *sql.DB, insertQueryTimeout, getQueryTimeout time.Duration) *User {
	return &User{
		db:                 db,
		insertQueryTimeout: insertQueryTimeout,
		getQueryTimeout:    getQueryTimeout,
	}
}

func (u *User) CreateUser(name, email, handle, password_hash string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.insertQueryTimeout)
	defer cancel()
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert user: %w", err)
	}
	_, err = u.db.ExecContext(
		ctx,
		"INSERT INTO users (id, name, email, handle, password_hash) VALUES (?, ?, ?, ?, ?)",
		id[:], name, email, handle, password_hash,
	)
	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
		if strings.Contains(mysqlError.Message, "email_unique") {
			return uuid.Nil, fmt.Errorf("insert user: %w", ErrEmailExists)
		} else if strings.Contains(mysqlError.Message, "handle_unique") {
			return uuid.Nil, fmt.Errorf("insert user: %w", ErrHandleExists)
		} else {
			return uuid.Nil, fmt.Errorf("insert user: unexprected unique constraint: %w", err)
		}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return uuid.Nil, fmt.Errorf("insert user: %w", ErrDatabaseTimeout)
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert user: %w", err)
	}

	return id, nil
}

func (u *User) GetAuthData(identifier string) (uuid.UUID, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), u.getQueryTimeout)
	defer cancel()
	var id uuid.UUID
	var passwordHash []byte
	err := u.db.QueryRowContext(
		ctx,
		fmt.Sprintf("SELECT id, password_hash FROM users WHERE %s = ?", getIdentiferType(identifier)),
		identifier,
	).Scan(&id, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, nil, fmt.Errorf("get user credentials: %w", ErrNoSuchUser)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return uuid.Nil, nil, fmt.Errorf("get user credentials: %w", ErrDatabaseTimeout)
	}
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("get user credentials: %w", err)
	}
	return id, passwordHash, nil
}

func getIdentiferType(identifier string) string {
	if strings.ContainsRune(identifier, '@') {
		return "email"
	}
	return "handle"
}
