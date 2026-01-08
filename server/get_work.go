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

// GetWork retrieves a work by UUID
func (s *Server) GetWork(ctx context.Context, request vcrest.GetWorkRequestObject) (outResp vcrest.GetWorkResponseObject, outErr error) {
	// Validate request.
	requestUuid, err := internal.AsUUID(request.Uuid)
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
			Uuid:  request.Uuid,
			Movie: movieBody.ToAPI(),
		}
		return
	case internal.WorkKindMovieEdition:
		var editionBody internal.MovieEditionWork
		if err := json.Unmarshal(bodyRaw, &editionBody); err != nil {
			outResp = vcrest.GetWork500JSONResponse{
				Message: fmt.Sprintf("failed to unmarshal movie edition work body: %v", err),
			}
			return
		}
		outResp = vcrest.GetWork200JSONResponse{
			Uuid:         request.Uuid,
			MovieEdition: editionBody.ToAPI(),
		}
		return
	default:
		outResp = vcrest.GetWork500JSONResponse{
			Message: fmt.Sprintf("unimplemented work kind: %s", kind),
		}
		return
	}
}
