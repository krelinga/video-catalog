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

// PutMovieEdition creates or updates a movie edition work with the given UUID
func (s *Server) PutMovieEdition(ctx context.Context, request vcrest.PutMovieEditionRequestObject) (outResp vcrest.PutMovieEditionResponseObject, _ error) {
	// Validate request.
	requestUuid, err := uuid.Parse(request.Uuid.String())
	if err != nil {
		outResp = vcrest.PutMovieEdition400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PutMovieEdition400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	var body internal.MovieEditionWork
	if !request.Body.EditionType.IsSpecified() || request.Body.EditionType.IsNull() || request.Body.EditionType.MustGet() == "" {
		outResp = vcrest.PutMovieEdition400JSONResponse{
			Message: "non-empty editionType is required",
		}
		return
	} else {
		body.EditionType = request.Body.EditionType.MustGet()
	}

	bodyRaw, err := json.Marshal(body)
	if err != nil {
		outResp = vcrest.PutMovieEdition500JSONResponse{
			Message: fmt.Sprintf("failed to marshal database body: %v", err),
		}
		return
	}

	row := s.Pool.QueryRow(ctx, `
		INSERT INTO works (uuid, kind, body)
		VALUES ($1, $2, $3)
		ON CONFLICT (uuid) DO UPDATE
		SET body = EXCLUDED.body
		WHERE works.kind = $2
		RETURNING xmax`,
		requestUuid, internal.WorkKindMovieEdition, bodyRaw)
	var xmax uint32
	if err := row.Scan(&xmax); errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.PutMovieEdition409JSONResponse{
			Message: "work with given UUID already exists with different kind",
		}
		return
	} else if err != nil {
		outResp = vcrest.PutMovieEdition500JSONResponse{
			Message: fmt.Sprintf("failed to insert/update work: %v", err),
		}
		return
	}

	if xmax == 0 {
		outResp = vcrest.PutMovieEdition201Response{}
	} else {
		outResp = vcrest.PutMovieEdition200Response{}
	}
	return
}
