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

// GetSource retrieves a source by UUID
func (s *Server) GetSource(ctx context.Context, request vcrest.GetSourceRequestObject) (outResp vcrest.GetSourceResponseObject, outErr error) {
	// Validate request.
	requestUuid, err := internal.ParseUUID(request.Uuid.String())
	if err != nil {
		outResp = vcrest.GetSource400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.GetSource500JSONResponse{
			Message: fmt.Sprintf("failed to begin transaction: %v", err),
		}
		return
	}
	defer txn.Rollback(ctx)

	var kind internal.SourceKind
	var bodyRaw json.RawMessage
	err = txn.QueryRow(ctx, `
		SELECT kind, body
		FROM sources
		WHERE uuid = $1
	`, requestUuid).Scan(&kind, &bodyRaw)
	if errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.GetSource404JSONResponse{
			Message: "source not found",
		}
		return
	} else if err != nil {
		outResp = vcrest.GetSource500JSONResponse{
			Message: fmt.Sprintf("failed to query source: %v", err),
		}
		return
	}

	if !kind.IsValid() {
		outResp = vcrest.GetSource500JSONResponse{
			Message: fmt.Sprintf("invalid source kind in database: %s", kind),
		}
		return
	}

	switch kind {
	case internal.SourceKindFile:
		var fileBody internal.FileSource
		if err := json.Unmarshal(bodyRaw, &fileBody); err != nil {
			outResp = vcrest.GetSource500JSONResponse{
				Message: fmt.Sprintf("failed to unmarshal file source body: %v", err),
			}
			return
		}
		outResp = vcrest.GetSource200JSONResponse{
			Uuid: request.Uuid,
			File: fileBody.ToAPI(),
		}
		return
	case internal.SourceKindDisc:
		var discBody internal.DiscSource
		if err := json.Unmarshal(bodyRaw, &discBody); err != nil {
			outResp = vcrest.GetSource500JSONResponse{
				Message: fmt.Sprintf("failed to unmarshal disc source body: %v", err),
			}
			return
		}
		outResp = vcrest.GetSource200JSONResponse{
			Uuid: request.Uuid,
			Disc: discBody.ToAPI(),
		}
		return
	default:
		outResp = vcrest.GetSource500JSONResponse{
			Message: fmt.Sprintf("unimplemented source kind: %s", kind),
		}
		return
	}
}
