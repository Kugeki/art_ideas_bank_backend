package domain

import (
	"io"
	"strings"
)

type File interface {
	io.ReadCloser
	io.Seeker
}

func IsTypeImageOrVideo(contentType string) bool {
	return strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/")
}
