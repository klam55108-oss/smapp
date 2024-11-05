package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	commonhttp "smapp/common/http"
	"smapp/post/repository"
	"smapp/post/service"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gorilla/mux"
)

func CreateLike(likeService *service.Like) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		entityType, ok := vars["entity_type"]
		if !ok {
			panic("create like: missing entity type")
		}
		if entityType != "posts" && entityType != "comments" {
			fmt.Println(entityType)
			panic("create like: invalid entity type")
		}
		entityID, ok := vars["entity_id"]
		if !ok {
			panic("create like: missing entity ID")
		}
		if err := validation.Validate(entityID, is.UUIDv4); err != nil {
			commonhttp.JSONError(w, fmt.Sprintf("Invalid entity ID: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// Headers set by the gateway
		authorID := r.Header.Get("X-User-Id")
		if authorID == "" {
			commonhttp.JSONError(w, "Missing X-User-Id header", http.StatusUnauthorized)
			return
		}

		err := likeService.Create(r.Context(), entityType, entityID, authorID)
		if errors.Is(err, repository.ErrPostDoesNotExist) {
			commonhttp.JSONError(w, "Post ID does not exist", http.StatusBadRequest)
			log.Println(err)
			return
		}
		if errors.Is(err, repository.ErrCommentDoesNotExist) {
			commonhttp.JSONError(w, "Comment ID does not exist", http.StatusBadRequest)
			log.Println(err)
			return
		}
		if errors.Is(err, repository.ErrLikeExists) {
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
