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

// PutDiscSource adds or updates a disc source with the given UUID
func (s *Server) PutDiscSource(ctx context.Context, request vcrest.PutDiscSourceRequestObject) (outResp vcrest.PutDiscSourceResponseObject, _ error) {
	// Validate request.
	requestUuid, err := uuid.Parse(request.Uuid.String())
	if err != nil {
		outResp = vcrest.PutDiscSource400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PutDiscSource400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	if !request.Body.OrigDirName.IsSpecified() || request.Body.OrigDirName.IsNull() || request.Body.OrigDirName.MustGet() == "" {
		outResp = vcrest.PutDiscSource400JSONResponse{
			Message: "non-empty OrigDirName is required",
		}
		return
	}
	if !request.Body.Path.IsSpecified() || request.Body.Path.IsNull() || request.Body.Path.MustGet() == "" {
		outResp = vcrest.PutDiscSource400JSONResponse{
			Message: "non-empty Path is required",
		}
		return
	}
	if request.Body.AllFilesAdded.IsSpecified() && request.Body.AllFilesAdded.IsNull() {
		outResp = vcrest.PutDiscSource400JSONResponse{
			Message: "AllFilesAdded must not be null",
		}
		return
	}

	body := internal.DiscSource{
		OrigDirName: request.Body.OrigDirName.MustGet(),
		Path:        request.Body.Path.MustGet(),
	}
	if request.Body.AllFilesAdded.IsSpecified() {
		body.AllFilesAdded = request.Body.AllFilesAdded.MustGet()
	}

	bodyRaw, err := json.Marshal(body)
	if err != nil {
		outResp = vcrest.PutDiscSource500JSONResponse{
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
		requestUuid, internal.SourceKindDisc, bodyRaw)
	var xmax uint32
	if err := row.Scan(&xmax); errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.PutDiscSource409JSONResponse{
			Message: "source with given UUID already exists with different kind",
		}
		return
	} else if err != nil {
		outResp = vcrest.PutDiscSource500JSONResponse{
			Message: fmt.Sprintf("failed to insert/update source: %v", err),
		}
		return
	}

	if xmax == 0 {
		outResp = vcrest.PutDiscSource201Response{}
	} else {
		outResp = vcrest.PutDiscSource200Response{}
	}
	return
}
