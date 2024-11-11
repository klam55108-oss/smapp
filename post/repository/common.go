package repository

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrRecordExists   = errors.New("record already exists")
	ErrPostIDNotFound = errors.New("id not found in posts table")
)

// tx operations may return sql.ErrTxDone if the context is done and the transaction rollback has already completed. Return a context error instead for clarity in the service layer
func changeErrIfCtxDone(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil && errors.Is(err, sql.ErrTxDone) {
		return ctxErr
	}
	return err
}
