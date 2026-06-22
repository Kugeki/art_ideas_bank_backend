package domain

import "fmt"

type Tag struct {
	ID     string
	UserID int
	Path   string
	Name   string
}

type TagError struct {
	ID   string
	Path string
	Err  error
}

func (e *TagError) Error() string {
	if len(e.Path) > 0 {
		return fmt.Sprintf("tag path(%s): %s", e.Path, e.Err.Error())
	}
	if len(e.ID) > 0 {
		return fmt.Sprintf("tag id(%s): %s", e.ID, e.Err.Error())
	}
	return fmt.Sprintf("tag error: %s", e.Err.Error())
}

func (e *TagError) Unwrap() error {
	return e.Err
}

func (e *TagError) With(err error) error {
	e.Err = err
	return e
}
