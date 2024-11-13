package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"smapp/common/jsonresp"
	"smapp/post/service"

	commonmw "smapp/common/middleware"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func CreateLike(likeService *service.Like) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entityID, err := uuid.Parse(mux.Vars(r)["entity_id"])
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("Invalid entity ID: %s", err.Error()), http.StatusBadRequest)
			return
		}

		authorID, err := commonmw.GetUserID(r.Context())
		if err != nil {
			log.Println(err)
			jsonresp.ErrorWithDefaultMessage(w, http.StatusInternalServerError)
			return
		}

		err = likeService.Create(r.Context(), entityID, authorID)
		if errors.Is(err, service.ErrPostNotFound) {
			jsonresp.Error(w, "Post ID does not exist", http.StatusBadRequest)
			log.Println(err)
			return
		}
		if errors.Is(err, service.ErrCommentNotFound) {
			jsonresp.Error(w, "Comment ID does not exist", http.StatusBadRequest)
			log.Println(err)
			return
		}
		if errors.Is(err, service.ErrLikeExists) {
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
