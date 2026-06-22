package domain

import (
	"fmt"
	"time"
)

type Image struct {
	ID         string
	UserID     int
	Extension  string
	S3Key      string
	UploadedAt time.Time

	Tags []Tag
}

type ImageError struct {
	ID  string
	Err error
}

func (e *ImageError) Error() string {
	if len(e.ID) > 0 {
		return fmt.Sprintf("image id(%s): %s", e.ID, e.Err.Error())
	}
	return fmt.Sprintf("image error: %s", e.Err.Error())
}

func (e *ImageError) Unwrap() error {
	return e.Err
}

func (e *ImageError) With(err error) error {
	e.Err = err
	return e
}
