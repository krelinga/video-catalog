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

// PatchDiscSource updates fields of a disc source with the given UUID
func (s *Server) PatchDiscSource(ctx context.Context, request vcrest.PatchDiscSourceRequestObject) (outResp vcrest.PatchDiscSourceResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.AsUUID(request.Uuid)
	if err != nil {
		outResp = vcrest.PatchDiscSource400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PatchDiscSource400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	if err := internal.FieldNotEmpty(request.Body.OrigDirName); err != nil {
		outResp = vcrest.PatchDiscSource400JSONResponse{
			Message: fmt.Sprintf("OrigDirName: %v", err),
		}
		return
	}
	origDirName := internal.FieldMay(request.Body.OrigDirName)

	if err := internal.FieldNotEmpty(request.Body.Path); err != nil {
		outResp = vcrest.PatchDiscSource400JSONResponse{
			Message: fmt.Sprintf("Path: %v", err),
		}
		return
	}
	path := internal.FieldMay(request.Body.Path)

	if err := internal.FieldNotNull(request.Body.AllFilesAdded); err != nil {
		outResp = vcrest.PatchDiscSource400JSONResponse{
			Message: fmt.Sprintf("AllFilesAdded: %v", err),
		}
		return
	}

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.PatchDiscSource500JSONResponse{
			Message: fmt.Sprintf("failed to begin transaction: %v", err),
		}
		return
	}
	defer txn.Rollback(ctx)

	var kind internal.SourceKind
	var rawBody json.RawMessage
	row := txn.QueryRow(ctx, `
		SELECT kind, body
		FROM sources
		WHERE uuid = $1
	`, requestUuid)
	err = row.Scan(&kind, &rawBody)
	if errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.PatchDiscSource404JSONResponse{
			Message: "source not found",
		}
		return
	} else if err != nil {
		outResp = vcrest.PatchDiscSource500JSONResponse{
			Message: fmt.Sprintf("failed to query source: %v", err),
		}
		return
	} else if kind != internal.SourceKindDisc {
		outResp = vcrest.PatchDiscSource409JSONResponse{
			Message: "source is not a disc",
		}
		return
	}
	var body internal.DiscSource
	if err := json.Unmarshal(rawBody, &body); err != nil {
		outResp = vcrest.PatchDiscSource500JSONResponse{
			Message: fmt.Sprintf("failed to unmarshal source body: %v", err),
		}
		return
	}

	if origDirName != nil {
		body.OrigDirName = *origDirName
	}
	if path != nil {
		body.Path = *path
	}
	internal.FieldSet(request.Body.AllFilesAdded, &body.AllFilesAdded)

	rawBody, err = json.Marshal(body)
	if err != nil {
		outResp = vcrest.PatchDiscSource500JSONResponse{
			Message: fmt.Sprintf("failed to marshal database body: %v", err),
		}
		return
	}

	_, err = txn.Exec(ctx, `
		UPDATE sources
		SET body = $2
		WHERE uuid = $1
	`, requestUuid, rawBody)
	if err != nil {
		outResp = vcrest.PatchDiscSource500JSONResponse{
			Message: fmt.Sprintf("failed to update source: %v", err),
		}
		return
	}

	if err := txn.Commit(ctx); err != nil {
		outResp = vcrest.PatchDiscSource500JSONResponse{
			Message: fmt.Sprintf("failed to commit transaction: %v", err),
		}
		return
	}

	outResp = vcrest.PatchDiscSource200Response{}
	return
}
