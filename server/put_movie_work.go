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
)

// PutMovieWork adds or updates a movie work with the given UUID
func (s *Server) PutMovieWork(ctx context.Context, request vcrest.PutMovieWorkRequestObject) (outResp vcrest.PutMovieWorkResponseObject, outErr error) {
	// Validate request.
	requestUuid, err := uuid.Parse(request.Uuid.String())
	if err != nil {
		outResp = vcrest.PutMovieWork400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PutMovieWork400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	if request.Body.Title == "" {
		outResp = vcrest.PutMovieWork400JSONResponse{
			Message: "non-empty title is required",
		}
		return
	}

	// Start transaction
	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.PutMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to begin transaction: %v", err),
		}
		return
	}
	defer txn.Rollback(ctx)

	// Get existing work if it exists
	var exists bool
	var kind internal.WorkKind
	var bodyRaw json.RawMessage
	err = txn.QueryRow(ctx, `
		SELECT kind, body
		FROM works
		WHERE uuid = $1
	`, requestUuid).Scan(&kind, &bodyRaw)
	if err == nil {
		exists = true
	} else if !errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.PutMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to query existing work: %v", err),
		}
		return
	}

	var body internal.MovieWork
	if exists {
		if !kind.IsValid() {
			outResp = vcrest.PutMovieWork500JSONResponse{
				Message: fmt.Sprintf("existing work has invalid kind: %q", kind),
			}
			return
		}
		if kind != internal.WorkKindMovie {
			outResp = vcrest.PutMovieWork409JSONResponse{
				Message: fmt.Sprintf("work kind mismatch: existing kind is %q, but request is for %q", kind, internal.WorkKindMovie),
			}
			return
		}
		if err := json.Unmarshal(bodyRaw, &body); err != nil {
			outResp = vcrest.PutMovieWork500JSONResponse{
				Message: fmt.Sprintf("failed to unmarshal existing work body: %v", err),
			}
			return
		}
	} else {
		kind = internal.WorkKindMovie
	}

	// Update fields
	body.Title = request.Body.Title
	body.ReleaseYear = request.Body.ReleaseYear
	body.TmdbId = request.Body.TmdbId

	// Marshal body
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		outResp = vcrest.PutMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to marshal work body: %v", err),
		}
		return
	}
	bodyRaw = json.RawMessage(bodyBytes)

	// Insert or update work
	if exists {
		_, err = txn.Exec(ctx, `
			UPDATE works
			SET body = $2
			WHERE uuid = $1
		`, requestUuid, bodyRaw)
		if err != nil {
			outResp = vcrest.PutMovieWork500JSONResponse{
				Message: fmt.Sprintf("failed to update existing work: %v", err),
			}
			return
		}
	} else {
		_, err = txn.Exec(ctx, `
			INSERT INTO works (uuid, kind, body)
			VALUES ($1, $2, $3)
		`, requestUuid, kind, bodyRaw)
		if err != nil {
			outResp = vcrest.PutMovieWork500JSONResponse{
				Message: fmt.Sprintf("failed to insert new work: %v", err),
			}
			return
		}
	}

	// Commit transaction
	if err := txn.Commit(ctx); err != nil {
		outResp = vcrest.PutMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to commit transaction: %v", err),
		}
		return
	}

	// Success
	if exists {
		outResp = vcrest.PutMovieWork200Response{}
	} else {
		outResp = vcrest.PutMovieWork201Response{}
	}

	return 
}
