package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	commonhttp "smapp/common/http"
	"smapp/user/repository"
	"smapp/user/service"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gorilla/mux"
)

func Follow(followService *service.Follow) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		followedID, ok := mux.Vars(r)["user_id"]
		if !ok {
			panic("follow: missing user ID")
		}
		if err := validation.Validate(followedID, is.UUIDv4); err != nil {
			commonhttp.JSONError(w, fmt.Sprintf("Invalid user ID: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// Headers set by the gateway
		followerID := r.Header.Get("X-User-Id")
		if followerID == "" {
			commonhttp.JSONError(w, "Missing X-User-Id header", http.StatusUnauthorized)
			return
		}

		err := followService.Create(r.Context(), followerID, followedID)
		if errors.Is(err, service.ErrSelfFollow) {
			commonhttp.JSONError(w, "Cannot follow self", http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrUserDoesNotExist) {
			commonhttp.JSONError(w, "User not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrFollowExists) {
			response := map[string]interface{}{"status": "unchanged"}
			commonhttp.JSONResponse(w, response, http.StatusOK)
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println(err)
			commonhttp.JSONErrorWithDefaultMessage(w, http.StatusRequestTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			// client disconnected
			log.Println(err)
			return
		}
		if err != nil {
			log.Println(err)
			commonhttp.JSONErrorWithDefaultMessage(w, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{"status": "success"}
		commonhttp.JSONResponse(w, response, http.StatusCreated)
	})
}
