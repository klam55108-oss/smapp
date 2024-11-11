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
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gorilla/mux"
)

type CreatePostRequestBody struct {
	Body   string                `json:"body"`
	Images []model.ImageLocation `json:"images"`
}

func (post *CreatePostRequestBody) Validate() error {
	return validation.ValidateStruct(
		post,
		validation.Field(
			&post.Body,
			validation.When(
				len(post.Images) == 0,
				validation.Required.Error("body or imageURLs is required"),
			),
			validation.Length(1, 5000),
		),
		validation.Field(
			&post.Images,
			validation.When(
				post.Body == "",
				validation.By(func(value interface{}) error {
					urls := value.([]model.ImageLocation)
					if len(urls) == 0 {
						return errors.New("body or imageURLs is required")
					}
					return nil
				}),
			),
			validation.Each(
				validation.By(func(value interface{}) error {
					image := value.(model.ImageLocation)
					return image.Validate()
				}),
			),
		),
	)
}

func CreatePost(postService *service.Post) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var post CreatePostRequestBody
		err := json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			jsonresp.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		err = post.Validate()
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

		// Headers set by the gateway
		authorID := r.Header.Get("X-User-Id")
		if authorID == "" {
			jsonresp.Error(w, "Missing X-User-Id header", http.StatusUnauthorized)
			return
		}

		id, err := postService.Create(r.Context(), post.Body, authorID, post.Images)
		if errors.Is(err, service.ErrInvalidImage) {
			jsonresp.Error(w, "One or more provided image locations are invalid or inaccessible", http.StatusBadRequest)
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

func GetPost(postService *service.Post) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID, ok := mux.Vars(r)["post_id"]
		if !ok {
			panic("get post: missing post ID")
		}
		err := validation.Validate(postID, is.UUIDv4)
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("Invalid post ID: %s", err.Error()), http.StatusBadRequest)
			return
		}

		post, err := postService.GetWithCounts(r.Context(), postID)
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
			"data":   map[string]interface{}{"post": post},
		}
		jsonresp.Response(w, response, http.StatusOK)
	})
}

func GetFeed(postService *service.Post) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Headers set by the gateway
		authorID := r.Header.Get("X-User-Id")
		if authorID == "" {
			jsonresp.Error(w, "Missing X-User-Id header", http.StatusUnauthorized)
			return
		}

		lastLoadedTimestamp, err := time.Parse(time.RFC3339, r.URL.Query().Get("last_loaded_timestamp"))
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("last_loaded_timestamp: should be in format %s", time.RFC3339), http.StatusBadRequest)
			return
		}
		lastLoadedID := r.URL.Query().Get("last_loaded_id")
		if err := validation.Validate(lastLoadedID, validation.Required, is.UUIDv4); err != nil {
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
		posts, nextCursor, err := postService.GetFeed(r.Context(), authorID, cursor, limit)
		if errors.Is(err, service.ErrPostsPaginationLimitInvalid) {
			jsonresp.Error(w, err.Error(), http.StatusBadRequest)
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
				"posts":       posts,
				"next_cursor": nextCursor,
			},
		}
		jsonresp.Response(w, response, http.StatusOK)
	}
}
