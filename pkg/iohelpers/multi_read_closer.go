package iohelpers

import (
	"errors"
	"io"
)

type MultiReadCloser struct {
	reader  io.Reader
	closers []io.Closer
}

func NewMultiReadCloser(readers ...io.ReadCloser) io.ReadCloser {
	ioReaders := make([]io.Reader, len(readers))
	closers := make([]io.Closer, len(readers))

	for i, r := range readers {
		ioReaders[i] = r
		closers[i] = r
	}

	return &MultiReadCloser{
		reader:  io.MultiReader(ioReaders...),
		closers: closers,
	}
}

func (m *MultiReadCloser) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

func (m *MultiReadCloser) Close() error {
	var errs []error
	for _, closer := range m.closers {
		if err := closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
