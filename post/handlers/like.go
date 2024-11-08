package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"smapp/common/jsonresp"
	"smapp/post/repository"
	"smapp/post/service"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gorilla/mux"
)

func CreateLike(likeService *service.Like) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entityID, ok := mux.Vars(r)["entity_id"]
		if !ok {
			panic("create like: missing entity ID")
		}
		if err := validation.Validate(entityID, is.UUIDv4); err != nil {
			jsonresp.Error(w, fmt.Sprintf("Invalid entity ID: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// Headers set by the gateway
		authorID := r.Header.Get("X-User-Id")
		if authorID == "" {
			jsonresp.Error(w, "Missing X-User-Id header", http.StatusUnauthorized)
			return
		}

		err := likeService.Create(r.Context(), entityID, authorID)
		if errors.Is(err, repository.ErrPostDoesNotExist) {
			jsonresp.Error(w, "Post ID does not exist", http.StatusBadRequest)
			log.Println(err)
			return
		}
		if errors.Is(err, repository.ErrCommentDoesNotExist) {
			jsonresp.Error(w, "Comment ID does not exist", http.StatusBadRequest)
			log.Println(err)
			return
		}
		if errors.Is(err, repository.ErrLikeExists) {
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
