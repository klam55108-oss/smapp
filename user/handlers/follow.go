package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"smapp/common/jsonresp"
	commonmw "smapp/common/middleware"
	"smapp/user/service"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func Follow(followService *service.Follow) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		followedID, err := uuid.Parse(mux.Vars(r)["user_id"])
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("Invalid user ID: %s", err.Error()), http.StatusBadRequest)
			return
		}

		followerID, err := commonmw.GetUserID(r.Context())
		if err != nil {
			log.Println(err)
			jsonresp.ErrorWithDefaultMessage(w, http.StatusInternalServerError)
			return
		}

		err = followService.Create(r.Context(), followerID, followedID)
		if errors.Is(err, service.ErrSelfFollow) {
			jsonresp.Error(w, "Cannot follow self", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrUserNotFound) {
			jsonresp.Error(w, "User not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrFollowExists) {
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
