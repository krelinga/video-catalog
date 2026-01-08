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

// PatchMovieEdition updates fields of a movie edition work with the given UUID
func (s *Server) PatchMovieEdition(ctx context.Context, request vcrest.PatchMovieEditionRequestObject) (outResp vcrest.PatchMovieEditionResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.ParseUUID(request.Uuid.String())
	if err != nil {
		outResp = vcrest.PatchMovieEdition400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PatchMovieEdition400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	newEditionType, updateEditionType, err := internal.ValidateOptionalNonEmptyString(request.Body.EditionType)
	if err != nil {
		outResp = vcrest.PatchMovieEdition400JSONResponse{
			Message: fmt.Sprintf("EditionType: %v", err),
		}
		return
	}

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.PatchMovieEdition500JSONResponse{
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
		outResp = vcrest.PatchMovieEdition404JSONResponse{
			Message: "work not found",
		}
		return
	} else if err != nil {
		outResp = vcrest.PatchMovieEdition500JSONResponse{
			Message: fmt.Sprintf("failed to query work: %v", err),
		}
		return
	} else if kind != internal.WorkKindMovieEdition {
		outResp = vcrest.PatchMovieEdition409JSONResponse{
			Message: "work is not a movie edition",
		}
		return
	}
	var body internal.MovieEditionWork
	if err := json.Unmarshal(rawBody, &body); err != nil {
		outResp = vcrest.PatchMovieEdition500JSONResponse{
			Message: fmt.Sprintf("failed to unmarshal work body: %v", err),
		}
		return
	}

	if updateEditionType {
		body.EditionType = newEditionType
	}

	rawBody, err = json.Marshal(body)
	if err != nil {
		outResp = vcrest.PatchMovieEdition500JSONResponse{
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
		outResp = vcrest.PatchMovieEdition500JSONResponse{
			Message: fmt.Sprintf("failed to update work: %v", err),
		}
		return
	}

	if err := txn.Commit(ctx); err != nil {
		outResp = vcrest.PatchMovieEdition500JSONResponse{
			Message: fmt.Sprintf("failed to commit transaction: %v", err),
		}
		return
	}

	outResp = vcrest.PatchMovieEdition200Response{}
	return
}
