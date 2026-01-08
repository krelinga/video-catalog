package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
)

// PatchMovieWork updates fields of a movie work with the given UUID
func (s *Server) PatchMovieWork(ctx context.Context, request vcrest.PatchMovieWorkRequestObject) (outResp vcrest.PatchMovieWorkResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.AsUUID(request.Uuid)
	if err != nil {
		outResp = vcrest.PatchMovieWork400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PatchMovieWork400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	if err := internal.FieldNotEmpty(request.Body.Title); err != nil {
		outResp = vcrest.PatchMovieWork400JSONResponse{
			Message: fmt.Sprintf("Title: %v", err),
		}
		return
	}
	title := internal.FieldMay(request.Body.Title)

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.PatchMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to begin transaction: %v", err),
		}
		return
	}
	defer txn.Rollback(ctx)

	var kind internal.WorkKind
	var rawBody json.RawMessage
	row := txn.QueryRow(ctx, `
		SELECT kind, body
		FROM works
		WHERE uuid = $1
	`, requestUuid)
	err = row.Scan(&kind, &rawBody)
	if errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.PatchMovieWork404JSONResponse{
			Message: "work not found",
		}
		return
	} else if err != nil {
		outResp = vcrest.PatchMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to query work: %v", err),
		}
		return
	} else if kind != internal.WorkKindMovie {
		outResp = vcrest.PatchMovieWork409JSONResponse{
			Message: "work is not a movie",
		}
		return
	}
	var body internal.MovieWork
	if err := json.Unmarshal(rawBody, &body); err != nil {
		outResp = vcrest.PatchMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to unmarshal work body: %v", err),
		}
		return
	}

	if title != nil {
		body.Title = *title
	}
	internal.FieldSetClear(request.Body.ReleaseYear, &body.ReleaseYear)
	internal.FieldSetClear(request.Body.TmdbId, &body.TmdbId)

	rawBody, err = json.Marshal(body)
	if err != nil {
		outResp = vcrest.PatchMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to marshal database body: %v", err),
		}
		return
	}

	_, err = txn.Exec(ctx, `
		UPDATE works
		SET body = $2
		WHERE uuid = $1
	`, requestUuid, rawBody)
	if err != nil {
		outResp = vcrest.PatchMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to update work: %v", err),
		}
		return
	}

	if err := txn.Commit(ctx); err != nil {
		outResp = vcrest.PatchMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to commit transaction: %v", err),
		}
		return
	}

	outResp = vcrest.PatchMovieWork200Response{}
	return
}
