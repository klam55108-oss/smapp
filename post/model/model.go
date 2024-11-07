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
	ID        string
	AuthorID  string
	Body      string
	Images    []ImageLocation
	CreatedAt time.Time
}

type ImageLocation struct {
	Bucket string
	Key    string
}

func (image *ImageLocation) Validate() error {
	return validation.ValidateStruct(
		image,
		validation.Field(&image.Bucket, validation.Required, validation.Length(1, 63)),
		validation.Field(&image.Key, validation.Required, validation.Length(1, 1024)),
	)
}
