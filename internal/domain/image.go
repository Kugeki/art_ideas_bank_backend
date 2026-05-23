package domain

import "time"

type Image struct {
	ID         string
	UserID     int
	S3Key      string
	UploadedAt time.Time

	Tags []Tag
}
