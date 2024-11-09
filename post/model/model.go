package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type EntityType string

const (
	PostType    EntityType = "posts"
	CommentType EntityType = "comments"
)

type Post struct {
	ID        string          `json:"id"`
	AuthorID  string          `json:"author_id"`
	Body      string          `json:"body"`
	Images    []ImageLocation `json:"images"`
	CreatedAt time.Time       `json:"created_at"`
}

type ImageLocation struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

func (image *ImageLocation) Validate() error {
	return validation.ValidateStruct(
		image,
		validation.Field(&image.Bucket, validation.Required, validation.Length(1, 63)),
		validation.Field(&image.Key, validation.Required, validation.Length(1, 1024)),
	)
}
