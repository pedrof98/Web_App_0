package service 

import (
	"errors"
	"fmt"

	"traffic-monitoring-go/internal/repository"
)

// common service errors
var (
	// ErrNotFound is returned when an entity is not found
	ErrNotFound = errors.New("resource not found")

	// ErrBadRequest is returned when the request is invalid
	ErrBadRequest = errors.New("invalid request")

	// ErrConflict is returned when a resource already exists
	ErrConflict = errors.New("resource already exists")

	// ErrForbidden is returned when a user doesn't have permission
	ErrForbidden = errors.New("operation not permitted")

	// ErrInternal is returned for internal server errors
	ErrInternal = erros.New("internal server error")
)

// WrapError maps repository errors to service errors
func WrapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, repository.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, repositoru.ErrDuplicate):
		return ErrConflict
	case errors.Is(err, repository.ErrForeignKey):
		return ErrBadRequest
	default:
		return fmt.Errorf("%w %v", ErrInternal, err)
	}
}


