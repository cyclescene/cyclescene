package apierror

import (
	"errors"
)

var (
	ErrTokenUsed        = errors.New("token used")
	ErrTokenExpired     = errors.New("token expired")
	ErrCityMismatch     = errors.New("city mismatch")
	ErrMissingFields    = errors.New("missing required field")
	ErrMissingOccurence = errors.New("at least one occurrence is required")
)
