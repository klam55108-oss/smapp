package model

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

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
