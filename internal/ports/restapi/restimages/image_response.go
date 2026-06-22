package restimages

import (
	"time"

	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi/resttags"
)

type ImageResp struct {
	ID         string             `json:"id"`
	Extension  string             `json:"extension"`
	UploadedAt time.Time          `json:"uploaded_at"`
	Tags       []resttags.TagResp `json:"tags"`
}
