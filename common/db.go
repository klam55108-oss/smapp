package common

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func WaitForDB(ctx context.Context, db *sql.DB) error {
	retryTime := 1 * time.Second
	for {
		if db.Ping() == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			if db.Ping() == nil {
				return nil
			}
			return fmt.Errorf("wait for DB timeout: %w", ctx.Err())
		case <-time.After(retryTime):
			retryTime *= 2
		}
	}
}
