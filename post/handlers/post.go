package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"smapp/common/jsonresp"
	commonmw "smapp/common/middleware"
	"smapp/post/model"
	"smapp/post/service"
	"strconv"
	"time"

	"smapp/common/validation"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type CreatePostRequestBody struct {
	Body   string                `json:"body"`
	Images []model.ImageLocation `json:"images"`
}

func (post *CreatePostRequestBody) Validate() error {
	return ozzo.ValidateStruct(
		post,
		ozzo.Field(
			&post.Body,
			ozzo.When(
				len(post.Images) == 0,
				ozzo.Required.Error("body or imageURLs is required"),
			),
			ozzo.Length(1, 5000),
		),
		ozzo.Field(
			&post.Images,
			ozzo.When(
				post.Body == "",
				ozzo.By(func(value interface{}) error {
					urls := value.([]model.ImageLocation)
					if len(urls) == 0 {
						return errors.New("body or imageURLs is required")
					}
					return nil
				}),
			),
			ozzo.Each(
				ozzo.By(func(value interface{}) error {
					image := value.(model.ImageLocation)
					return image.Validate()
				}),
			),
		),
	)
}

func CreatePost(validator validation.Validator, postService service.Post) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var post CreatePostRequestBody
		err := json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			jsonresp.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		err = validator.Validate(&post)
		if errs, ok := err.(ozzo.Errors); ok {
			errs := errs.Filter().(ozzo.Errors)
			response := map[string]interface{}{
				"status":  "error",
				"message": "Validation failed",
				"errors":  errs,
			}
			jsonresp.Response(w, response, http.StatusBadRequest)
			return
		}
		if err != nil {
			log.Println(err)
			jsonresp.ErrorWithDefaultMessage(w, http.StatusInternalServerError)
			return
		}

		authorID, err := commonmw.GetUserID(r.Context())
		if err != nil {
			log.Println(err)
			jsonresp.ErrorWithDefaultMessage(w, http.StatusInternalServerError)
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

func GetPost(postService service.Post) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID, err := uuid.Parse(mux.Vars(r)["post_id"])
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

func GetFeed(postService service.Post) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorID, err := commonmw.GetUserID(r.Context())
		if err != nil {
			log.Println(err)
			jsonresp.ErrorWithDefaultMessage(w, http.StatusInternalServerError)
			return
		}

		lastLoadedTimestamp, err := time.Parse(time.RFC3339, r.URL.Query().Get("last_loaded_timestamp"))
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("last_loaded_timestamp: should be in format %s", time.RFC3339), http.StatusBadRequest)
			return
		}
		lastLoadedID, err := uuid.Parse(r.URL.Query().Get("last_loaded_id"))
		if err != nil {
			jsonresp.Error(w, fmt.Sprintf("Invalid last_loaded_id: %s", err.Error()), http.StatusBadRequest)
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
