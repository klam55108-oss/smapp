package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	commonhttp "smapp/common/http"
	"smapp/post/repository"
	"smapp/post/service"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type CreateCommentRequestBody struct {
	PostID string `json:"post_id"`
	Body   string `json:"body"`
}

func (comment *CreateCommentRequestBody) Validate() error {
	return validation.ValidateStruct(
		comment,
		validation.Field(&comment.PostID, validation.Required, is.UUIDv4),
		validation.Field(&comment.Body, validation.Required, validation.Length(1, 5000)),
	)
}

func CreateComment(commentService *service.Comment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var comment CreateCommentRequestBody
		err := json.NewDecoder(r.Body).Decode(&comment)
		if err != nil {
			commonhttp.JSONError(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		err = comment.Validate()
		if err != nil {
			if e, ok := err.(validation.InternalError); ok {
				log.Println(e.InternalError())
				commonhttp.JSONErrorWithDefaultMessage(w, http.StatusInternalServerError)
				return
			}
			errors := (err.(validation.Errors).Filter()).(validation.Errors)
			commonhttp.JSONValidationError(w, errors, http.StatusBadRequest)
			return
		}

		// Headers set by the gateway
		authorID := r.Header.Get("X-User-Id")
		if authorID == "" {
			commonhttp.JSONError(w, "Missing X-User-Id header", http.StatusUnauthorized)
			return
		}

		id, err := commentService.Create(r.Context(), comment.PostID, authorID, comment.Body)
		if errors.Is(err, repository.ErrPostDoesNotExist) {
			commonhttp.JSONError(w, "Post ID does not exist", http.StatusBadRequest)
			log.Println(err)
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

		response := map[string]interface{}{"id": id}
		commonhttp.JSONResponse(w, response, http.StatusCreated)
	})
}
