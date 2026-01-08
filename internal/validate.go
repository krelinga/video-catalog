package internal

import (
	"errors"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Sentinel errors for validation functions.
// Callers should use these to construct field-specific error messages.
var (
	ErrRequired    = errors.New("is required")
	ErrInvalidUUID = errors.New("invalid UUID format")
	ErrEmpty       = errors.New("cannot be empty")
	ErrNull        = errors.New("cannot be null")
	ErrNullOrEmpty = errors.New("cannot be null or empty")
)

// ParseUUID parses a UUID string and returns ErrInvalidUUID if parsing fails.
func ParseUUID(uuidStr string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil, ErrInvalidUUID
	}
	return parsed, nil
}

// ValidateRequiredNullableUUID validates a nullable UUID field that is required.
// Returns ErrRequired if not specified or null, ErrInvalidUUID if the format is invalid,
// or ErrEmpty if the UUID is the nil UUID.
func ValidateRequiredNullableUUID(field nullable.Nullable[openapi_types.UUID]) (uuid.UUID, error) {
	if !field.IsSpecified() || field.IsNull() {
		return uuid.Nil, ErrRequired
	}
	parsed, err := uuid.Parse(field.MustGet().String())
	if err != nil {
		return uuid.Nil, ErrInvalidUUID
	}
	if parsed == uuid.Nil {
		return uuid.Nil, ErrEmpty
	}
	return parsed, nil
}

// ValidateOptionalNonNullableUUID validates an optional nullable UUID field that cannot be null if specified.
// Returns (uuid, true, nil) if specified and valid, (uuid.Nil, false, nil) if not specified,
// or an error (ErrNull, ErrInvalidUUID, ErrEmpty) if specified but invalid.
func ValidateOptionalNonNullableUUID(field nullable.Nullable[openapi_types.UUID]) (uuid.UUID, bool, error) {
	if !field.IsSpecified() {
		return uuid.Nil, false, nil
	}
	if field.IsNull() {
		return uuid.Nil, false, ErrNull
	}
	parsed, err := uuid.Parse(field.MustGet().String())
	if err != nil {
		return uuid.Nil, false, ErrInvalidUUID
	}
	if parsed == uuid.Nil {
		return uuid.Nil, false, ErrEmpty
	}
	return parsed, true, nil
}

// ValidateRequiredString validates a nullable string field that is required and non-empty.
// Returns ErrRequired if not specified or null, or ErrEmpty if the string is empty.
func ValidateRequiredString(field nullable.Nullable[string]) (string, error) {
	if !field.IsSpecified() || field.IsNull() {
		return "", ErrRequired
	}
	value := field.MustGet()
	if value == "" {
		return "", ErrEmpty
	}
	return value, nil
}

// ValidateOptionalNonEmptyString validates an optional nullable string field that cannot be null or empty if specified.
// Returns (value, true, nil) if specified and valid, ("", false, nil) if not specified,
// or ErrNullOrEmpty if specified but null or empty.
func ValidateOptionalNonEmptyString(field nullable.Nullable[string]) (string, bool, error) {
	if !field.IsSpecified() {
		return "", false, nil
	}
	if field.IsNull() {
		return "", false, ErrNullOrEmpty
	}
	value := field.MustGet()
	if value == "" {
		return "", false, ErrNullOrEmpty
	}
	return value, true, nil
}
