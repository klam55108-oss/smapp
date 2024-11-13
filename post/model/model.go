package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

type EntityType string

const (
	PostType    EntityType = "posts"
	CommentType EntityType = "comments"
)

type Post struct {
	ID           uuid.UUID       `json:"id"`
	AuthorID     uuid.UUID       `json:"author_id"`
	Body         string          `json:"body"`
	Images       []ImageLocation `json:"images"`
	CreatedAt    time.Time       `json:"created_at"`
	CommentCount *uint32         `json:"comment_count,omitempty"`
	LikeCount    *uint32         `json:"like_count,omitempty"`
}

type Comment struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"post_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	LikeCount *uint32   `json:"like_count,omitempty"`
}

type Cursor struct {
	LastLoadedTimestamp time.Time `json:"last_loaded_timestamp"`
	LastLoadedID        uuid.UUID `json:"last_loaded_id"`
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
