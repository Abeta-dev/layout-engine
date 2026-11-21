package domain

import "errors"

// Sentinel errors mapped by usecases and HTTP handlers to structured API responses.
var (
	ErrNotFound            = errors.New("resource not found")
	ErrConflict            = errors.New("resource already exists")
	ErrValidation          = errors.New("validation failed")
	ErrNoPublishedSnapshot = errors.New("layout has no published snapshot")
)
