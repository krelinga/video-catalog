package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
	"github.com/oapi-codegen/nullable"
)

// GetWork retrieves a work by UUID
func (s *Server) GetWork(ctx context.Context, request vcrest.GetWorkRequestObject) (outResp vcrest.GetWorkResponseObject, outErr error) {
	// Validate request.
	requestUuid, err := uuid.Parse(request.Uuid.String())
	if err != nil {
		outResp = vcrest.GetWork400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.GetWork500JSONResponse{
			Message: fmt.Sprintf("failed to begin transaction: %v", err),
		}
		return
	}
	defer txn.Rollback(ctx)

	var kind internal.WorkKind
	var bodyRaw json.RawMessage
	err = txn.QueryRow(ctx, `
		SELECT kind, body
		FROM works
		WHERE uuid = $1
	`, requestUuid).Scan(&kind, &bodyRaw)
	if errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.GetWork404JSONResponse{
			Message: "work not found",
		}
		return
	} else if err != nil {
		outResp = vcrest.GetWork500JSONResponse{
			Message: fmt.Sprintf("failed to query work: %v", err),
		}
		return
	}

	if !kind.IsValid() {
		outResp = vcrest.GetWork500JSONResponse{
			Message: fmt.Sprintf("invalid work kind in database: %s", kind),
		}
		return
	}

	switch kind {
	case internal.WorkKindMovie:
		var movieBody internal.MovieWork
		if err := json.Unmarshal(bodyRaw, &movieBody); err != nil {
			outResp = vcrest.GetWork500JSONResponse{
				Message: fmt.Sprintf("failed to unmarshal movie work body: %v", err),
			}
			return
		}
		outResp = vcrest.GetWork200JSONResponse{
			Uuid: request.Uuid,
			Movie: &vcrest.Movie{
				Title:       required(movieBody.Title),	
				ReleaseYear: optional(movieBody.ReleaseYear),
				TmdbId:      optional(movieBody.TmdbId),
			},
		}
		return
	default:
		outResp = vcrest.GetWork500JSONResponse{
			Message: fmt.Sprintf("unimplemented work kind: %s", kind),
		}
		return
	}
}

func required[T any](t T) nullable.Nullable[T] {
	return nullable.NewNullableWithValue(t)
}

func optional[T any](t *T) nullable.Nullable[T] {
	if t == nil {
		return nullable.Nullable[T]{}
	}
	return nullable.NewNullableWithValue(*t)
}