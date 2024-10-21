package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	commonhttp "smapp/common/http"
	"smapp/post_service/service"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type CreatePostRequestBody struct {
	Body      string   `json:"body"`
	AuthorID  string   `json:"authorID"`
	ImageURLs []string `json:"imageURLs"`
}

func (post *CreatePostRequestBody) Validate() error {
	return validation.ValidateStruct(
		post,
		validation.Field(
			&post.Body,
			validation.When(
				len(post.ImageURLs) == 0,
				validation.Required.Error("body or imageURLs is required"),
			),
			validation.Length(1, 5000),
		),
		validation.Field(&post.AuthorID, validation.Required, is.UUIDv4),
		validation.Field(
			&post.ImageURLs,
			validation.When(
				post.Body == "",
				validation.By(func(value interface{}) error {
					urls := value.([]string)
					if len(urls) == 0 {
						return errors.New("body or imageURLs is required")
					}
					return nil
				}),
			),
			validation.Each(is.URL),
		),
	)
}

func CreatePost(postService *service.Post) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var post CreatePostRequestBody
		err := json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			commonhttp.JSONError(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		err = post.Validate()
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

		id, err := postService.CreatePost(r.Context(), post.Body, post.AuthorID, post.ImageURLs)
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
