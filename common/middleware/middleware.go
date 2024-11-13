package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"smapp/common/jsonresp"
	"time"

	"github.com/google/uuid"
)

func WithRequestContextTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type userIDKey struct{}

func ParseUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Header set by the gateway
		userID, err := uuid.Parse(r.Header.Get("X-User-Id"))
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("Invalid X-User-Id header: %s", err.Error()), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// The error can be non-nil only if the ParseUserID middleware was not used, which is a bug.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(userIDKey{}).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context; ParseUserID middleware should be used")
	}
	return userID, nil
}
