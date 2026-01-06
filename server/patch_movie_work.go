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

// PatchMovieWork updates fields of a movie work with the given UUID
func (s *Server) PatchMovieWork(ctx context.Context, request vcrest.PatchMovieWorkRequestObject) (outResp vcrest.PatchMovieWorkResponseObject, _ error) {
	// Validate request.
	requestUuid, err := uuid.Parse(request.Uuid.String())
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
	if request.Body.Title.IsSpecified() && (request.Body.Title.IsNull() || request.Body.Title.MustGet() == "") {
		outResp = vcrest.PatchMovieWork400JSONResponse{
			Message: "cannot update title to be nil or empty",
		}
		return
	}

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

	if request.Body.Title.IsSpecified() {
		body.Title = request.Body.Title.MustGet()
	}
	if request.Body.ReleaseYear.IsSpecified() {
		if request.Body.ReleaseYear.IsNull() {
			body.ReleaseYear = nil
		} else {
			ry := request.Body.ReleaseYear.MustGet()
			body.ReleaseYear = &ry
		}
	}
	if request.Body.TmdbId.IsSpecified() {
		if request.Body.TmdbId.IsNull() {
			body.TmdbId = nil
		} else {
			tid := request.Body.TmdbId.MustGet()
			body.TmdbId = &tid
		}
	}

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