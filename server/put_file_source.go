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

// PutFileSource adds or updates a file source with the given UUID
func (s *Server) PutFileSource(ctx context.Context, request vcrest.PutFileSourceRequestObject) (outResp vcrest.PutFileSourceResponseObject, _ error) {
	// Validate request.
	requestUuid, err := uuid.Parse(request.Uuid.String())
	if err != nil {
		outResp = vcrest.PutFileSource400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PutFileSource400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	if !request.Body.Path.IsSpecified() || request.Body.Path.IsNull() || request.Body.Path.MustGet() == "" {
		outResp = vcrest.PutFileSource400JSONResponse{
			Message: "non-empty Path is required",
		}
		return
	}

	body := internal.FileSource{
		Path: request.Body.Path.MustGet(),
	}

	bodyRaw, err := json.Marshal(body)
	if err != nil {
		outResp = vcrest.PutFileSource500JSONResponse{
			Message: fmt.Sprintf("failed to marshal database body: %v", err),
		}
		return
	}

	row := s.Pool.QueryRow(ctx, `
		INSERT INTO sources (uuid, kind, body)
		VALUES ($1, $2, $3)
		ON CONFLICT (uuid) DO UPDATE
		SET body = EXCLUDED.body
		WHERE sources.kind = $2
		RETURNING xmax`,
		requestUuid, internal.SourceKindFile, bodyRaw)
	var xmax uint32
	if err := row.Scan(&xmax); errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.PutFileSource409JSONResponse{
			Message: "source with given UUID already exists with different kind",
		}
		return
	} else if err != nil {
		outResp = vcrest.PutFileSource500JSONResponse{
			Message: fmt.Sprintf("failed to insert/update source: %v", err),
		}
		return
	}

	if xmax == 0 {
		outResp = vcrest.PutFileSource201Response{}
	} else {
		outResp = vcrest.PutFileSource200Response{}
	}
	return
}