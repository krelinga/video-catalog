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

// FieldRequired checks that the field is specified.
func FieldRequired[T any](field nullable.Nullable[T]) error {
	if !field.IsSpecified() {
		return ErrRequired
	}
	return nil
}

// FieldNotNull checks that the field is not null.
// If the field is not specified, this returns nil.
func FieldNotNull[T any](field nullable.Nullable[T]) error {
	if field.IsSpecified() && field.IsNull() {
		return ErrNull
	}
	return nil
}

// FieldNotEmpty checks that the field is not empty.
// If the field is not specified or is null, this returns nil.
func FieldNotEmpty[T ~string](field nullable.Nullable[T]) error {
	if field.IsSpecified() && !field.IsNull() && field.MustGet() == "" {
		return ErrEmpty
	}
	return nil
}

// FieldValidUUID checks that the field is a valid UUID string.
// If the field is not specified or is null, this returns nil.
func FieldValidUUID[T fmt.Stringer](field nullable.Nullable[T]) error {
	if field.IsSpecified() && !field.IsNull() {
		_, err := AsUUID(field.MustGet())
		if err != nil {
			return ErrInvalidUUID
		}
	}
	return nil
}

// FieldNonZeroUUID checks that the field is a valid, non-zero UUID string.
// If the field is not specified or is null, this returns nil.
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

// FieldMustUUID parses and returns the UUID from the field.
// Panics if the field is not a valid UUID.
func FieldMustUUID[T fmt.Stringer](field nullable.Nullable[T]) uuid.UUID {
	parsed, err := AsUUID(field.MustGet())
	if err != nil {
		panic("FieldMustUUID called on invalid UUID: " + err.Error())
	}
	return parsed
}

// FieldMayUUID parses and returns the UUID contained in the field as a pointer if it is set and non-null.
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

// FieldMay returns a pointer to the value contained in the field if it is set and non-null.
func FieldMay[T any](field nullable.Nullable[T]) *T {
	if !field.IsSpecified() || field.IsNull() {
		return nil
	}
	val := field.MustGet()
	return &val
}

// AsUUID parses the given stringer as a UUID, returning an error if the format is invalid.
func AsUUID(in fmt.Stringer) (uuid.UUID, error) {
	parsed, err := uuid.Parse(in.String())
	if err != nil {
		return uuid.Nil, ErrInvalidUUID
	}
	return parsed, nil
}

// FieldSetClear sets the output pointer to nil if the field is null, or to the value if not null.
// Does nothing if the field is not specified.
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

// FieldSet sets the output value to the value contained in the field.
// Does nothing if the field is not specified or is null.
func FieldSet[T any](field nullable.Nullable[T], out *T) {
	if !field.IsSpecified() || field.IsNull() {
		return
	}
	*out = field.MustGet()
}

// FieldSetPtr sets the output pointer to point to the value contained in the field.
// Does nothing if the field is not specified or is null.
func FieldSetPtr[T any](field nullable.Nullable[T], out **T) {
	if !field.IsSpecified() || field.IsNull() {
		return
	}
	val := field.MustGet()
	*out = &val
}