package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"smapp/common/jsonresp"
	"smapp/post/model"
	"smapp/post/service"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type CreateCommentRequestBody struct {
	Body string `json:"body"`
}

func (comment *CreateCommentRequestBody) Validate() error {
	return validation.ValidateStruct(
		comment,
		validation.Field(&comment.Body, validation.Required, validation.Length(1, 5000)),
	)
}

func CreateComment(commentService *service.Comment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var comment CreateCommentRequestBody
		err := json.NewDecoder(r.Body).Decode(&comment)
		if err != nil {
			jsonresp.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		err = comment.Validate()
		if err != nil {
			if e, ok := err.(validation.InternalError); ok {
				log.Println(e.InternalError())
				jsonresp.ErrorWithDefaultMessage(w, http.StatusInternalServerError)
				return
			}
			errors := (err.(validation.Errors).Filter()).(validation.Errors)
			jsonresp.ValidationError(w, errors, http.StatusBadRequest)
			return
		}

		postID, err := uuid.Parse(mux.Vars(r)["post_id"])
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("Invalid post ID: %s", err.Error()), http.StatusBadRequest)
			return
		}

		// Headers set by the gateway
		authorID, err := uuid.Parse(r.Header.Get("X-User-Id"))
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("Invalid X-User-Id header: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		id, err := commentService.Create(r.Context(), postID, authorID, comment.Body)
		if errors.Is(err, service.ErrPostNotFound) {
			jsonresp.Error(w, "Post ID does not exist", http.StatusBadRequest)
			log.Println(err)
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

		response := map[string]interface{}{
			"status": "success",
			"id":     id,
		}
		jsonresp.Response(w, response, http.StatusCreated)
	})
}

func GetComments(commentService *service.Comment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID, err := uuid.Parse(mux.Vars(r)["post_id"])
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("Invalid post ID: %s", err.Error()), http.StatusBadRequest)
			return
		}

		lastLoadedTimestamp, err := time.Parse(time.RFC3339, r.URL.Query().Get("last_loaded_timestamp"))
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("last_loaded_timestamp: should be in format %s", time.RFC3339), http.StatusBadRequest)
			return
		}
		lastLoadedID, err := uuid.Parse(r.URL.Query().Get("last_loaded_id"))
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("invalid last_loaded_id: %s", err.Error()), http.StatusBadRequest)
			return
		}
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			jsonresp.Error(w, "limit: should be an integer", http.StatusBadRequest)
			return
		}

		cursor := model.Cursor{
			LastLoadedTimestamp: lastLoadedTimestamp,
			LastLoadedID:        lastLoadedID,
		}
		comments, nextCursor, err := commentService.GetPaginatedWithLikeCount(r.Context(), postID, cursor, limit)
		if errors.Is(err, service.ErrCommentsPaginationLimitInvalid) {
			jsonresp.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrPostNotFound) {
			jsonresp.Error(w, "Post not found", http.StatusNotFound)
			log.Println(err)
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

		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"comments":    comments,
				"next_cursor": nextCursor,
			},
		}
		jsonresp.Response(w, response, http.StatusOK)
	})
}
