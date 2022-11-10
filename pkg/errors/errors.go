package errors

import "errors"

// Defines base errors for other packages to use.
var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotFound        = errors.New("not found")
	ErrAlreadyExists   = errors.New("already exists")
)

// New wraps for errors.New function.
var New = errors.New //nolint:gochecknoglobals

// As wraps for errors.As function.
var As = errors.As //nolint:gochecknoglobals

// Is wraps for errors.Is function.
var Is = errors.Is //nolint:gochecknoglobals

// Unwrap wraps for errors.Unwrap function.
var Unwrap = errors.Unwrap //nolint:gochecknoglobals
