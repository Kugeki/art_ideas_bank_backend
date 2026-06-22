package domain

import "errors"

// the "rest" comment indicates that the error has a mapping to the rest api status. See project/internal/ports/rest/error_helpers.go
var (
	ErrNotFound      = errors.New("not found")      // rest
	ErrAlreadyExists = errors.New("already exists") // rest

	ErrEmptyTagPath              = errors.New("tag path cannot be empty")      // rest
	ErrTagIncorrectName          = errors.New("incorrect name")                // rest
	ErrSomeTagsNotFound          = errors.New("some tags not found")           // rest
	ErrTagHasAssociatedImages    = errors.New("has associated images")         // rest
	ErrCannotMoveTagToDescendant = errors.New("cannot move tag to descendant") // rest

	ErrLimitEqualOrLessThanZero = errors.New("limit equal or less than zero") // rest

	ErrWrongCredentials = errors.New("wrong credentials") // rest
	ErrInvalidToken     = errors.New("invalid token")     // rest

	ErrIncorrectContentType = errors.New("incorrect content type (need image or video)") // rest
)
