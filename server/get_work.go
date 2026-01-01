package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
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
	if err != nil {
		outResp = vcrest.GetWork404JSONResponse{
			Message: "work not found",
		}
		return
	}

	if !kind.IsValid() {
		outResp = vcrest.GetWork500JSONResponse{
			Message: fmt.Sprintf("invalid work kind in database: %s", kind),
		}
		return
	}

	rows, err := txn.Query(ctx, `
		SELECT source_uuid
		FROM works
		WHERE uuid = $1
	`, requestUuid)
	if err != nil {
		outResp = vcrest.GetWork500JSONResponse{
			Message: fmt.Sprintf("failed to query work sources: %v", err),
		}
		return
	}
	var sourcekUuid pgtype.UUID
	var sourceUuids []uuid.UUID
	_, err = pgx.ForEachRow(rows, []any{&sourcekUuid}, func() error {
		sourceUuids = append(sourceUuids, sourcekUuid.Bytes)
		return nil
	})
	if err != nil {
		outResp = vcrest.GetWork500JSONResponse{
			Message: fmt.Sprintf("failed to scan work sources: %v", err),
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
				Title:       movieBody.Title,
				ReleaseYear: movieBody.ReleaseYear,
				TmdbId:      movieBody.TmdbId,
			},
			SourceUuids: sourceUuids,
		}
		return
	default:
		outResp = vcrest.GetWork500JSONResponse{
			Message: fmt.Sprintf("unimplemented work kind: %s", kind),
		}
		return
	}
}