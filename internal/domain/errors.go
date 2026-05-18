package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrOrderWasCompleted = errors.New("order was completed, cant change")
)
