package internal

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
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

func FieldRequired[T any](field nullable.Nullable[T]) error {
	if !field.IsSpecified() {
		return ErrRequired
	}
	return nil
}

func FieldNotNull[T any](field nullable.Nullable[T]) error {
	if field.IsSpecified() && field.IsNull() {
		return ErrNull
	}
	return nil
}

func FieldNotEmpty[T ~string](field nullable.Nullable[T]) error {
	if field.IsSpecified() && !field.IsNull() && field.MustGet() == "" {
		return ErrEmpty
	}
	return nil
}

func FieldValidUUID[T fmt.Stringer](field nullable.Nullable[T]) error {
	if field.IsSpecified() && !field.IsNull() {
		_, err := AsUUID(field.MustGet())
		if err != nil {
			return ErrInvalidUUID
		}
	}
	return nil
}

func FieldNonZeroUUID[T fmt.Stringer](field nullable.Nullable[T]) error {
	if field.IsSpecified() && !field.IsNull() {
		parsed, err := AsUUID(field.MustGet())
		if err != nil {
			return ErrInvalidUUID
		}
		if parsed == uuid.Nil {
			return ErrEmpty
		}
	}
	return nil
}

func FieldMustUUID[T fmt.Stringer](field nullable.Nullable[T]) uuid.UUID {
	parsed, err := AsUUID(field.MustGet())
	if err != nil {
		panic("FieldMustUUID called on invalid UUID: " + err.Error())
	}
	return parsed
}

func FieldMayUUID[T fmt.Stringer](field nullable.Nullable[T]) *uuid.UUID {
	if !field.IsSpecified() || field.IsNull() {
		return nil
	}
	parsed, err := AsUUID(field.MustGet())
	if err != nil {
		panic("FieldMayUUID called on invalid UUID: " + err.Error())
	}
	return &parsed
}

func FieldMay[T any](field nullable.Nullable[T]) *T {
	if !field.IsSpecified() || field.IsNull() {
		return nil
	}
	val := field.MustGet()
	return &val
}

func AsUUID(in fmt.Stringer) (uuid.UUID, error) {
	parsed, err := uuid.Parse(in.String())
	if err != nil {
		return uuid.Nil, ErrInvalidUUID
	}
	return parsed, nil
}

func FieldSetClear[T any](field nullable.Nullable[T], out **T) {
	if !field.IsSpecified() {
		return
	}
	if field.IsNull() {
		*out = nil
	} else {
		val := field.MustGet()
		*out = &val
	}
}

func FieldSet[T any](field nullable.Nullable[T], out *T) {
	if !field.IsSpecified() || field.IsNull() {
		return
	}
	*out = field.MustGet()
}