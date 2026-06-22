package stdcontype

import (
	"bytes"
	"errors"
	"github.com/Kugeki/art_ideas_bank_backend/internal/domain"
	"github.com/Kugeki/art_ideas_bank_backend/pkg/iohelpers"
	"io"
	"mime"
	"net/http"
)

type StdContentTypeDetector struct {
}

func New() *StdContentTypeDetector {
	return &StdContentTypeDetector{}
}

func (s *StdContentTypeDetector) DetectFileContentType(src domain.File) (contentType string, newSrc domain.File, err error) {
	detectedType, _, err := s.detectContentType(src)
	if err != nil {
		return "", nil, err
	}

	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		return "", nil, err
	}

	return detectedType, src, nil
}

func (s *StdContentTypeDetector) DetectReaderContentType(src io.ReadCloser) (contentType string, newSrc io.ReadCloser, err error) {
	detectedType, head, err := s.detectContentType(src)
	if err != nil {
		return "", nil, err
	}

	return detectedType, iohelpers.NewMultiReadCloser(io.NopCloser(bytes.NewReader(head)), src), nil
}

func (s *StdContentTypeDetector) GetExtensionContentType(ext string) (contentType string, err error) {
	ct := mime.TypeByExtension(ext)
	if len(ct) <= 0 {
		return ct, errors.New("can't detect content type")
	}

	return ct, nil
}

func (s *StdContentTypeDetector) detectContentType(src io.ReadCloser) (contentType string, head []byte, err error) {
	buf := make([]byte, 512)
	n, err := io.ReadFull(src, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", nil, err
	}

	head = buf[:n]
	return http.DetectContentType(head), head, nil
}
