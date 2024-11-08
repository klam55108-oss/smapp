package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"smapp/common/jsonresp"
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
			jsonresp.Error(w, fmt.Sprintf("Invalid user ID: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// Headers set by the gateway
		followerID := r.Header.Get("X-User-Id")
		if followerID == "" {
			jsonresp.Error(w, "Missing X-User-Id header", http.StatusUnauthorized)
			return
		}

		err := followService.Create(r.Context(), followerID, followedID)
		if errors.Is(err, service.ErrSelfFollow) {
			jsonresp.Error(w, "Cannot follow self", http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrUserDoesNotExist) {
			jsonresp.Error(w, "User not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrFollowExists) {
			response := map[string]interface{}{"status": "unchanged"}
			jsonresp.Response(w, response, http.StatusOK)
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println(err)
			jsonresp.ErrorWithDefaultMessage(w, http.StatusRequestTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			// client disconnected
			log.Println(err)
			return
		}
		if err != nil {
			log.Println(err)
			jsonresp.ErrorWithDefaultMessage(w, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{"status": "success"}
		jsonresp.Response(w, response, http.StatusCreated)
	})
}
