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

// PatchFileSource updates fields of a file source with the given UUID
func (s *Server) PatchFileSource(ctx context.Context, request vcrest.PatchFileSourceRequestObject) (outResp vcrest.PatchFileSourceResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.ParseUUID(request.Uuid.String())
	if err != nil {
		outResp = vcrest.PatchFileSource400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PatchFileSource400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	newPath, updatePath, err := internal.ValidateOptionalNonEmptyString(request.Body.Path)
	if err != nil {
		outResp = vcrest.PatchFileSource400JSONResponse{
			Message: fmt.Sprintf("Path: %v", err),
		}
		return
	}

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.PatchFileSource500JSONResponse{
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
		outResp = vcrest.PatchFileSource404JSONResponse{
			Message: "source not found",
		}
		return
	} else if err != nil {
		outResp = vcrest.PatchFileSource500JSONResponse{
			Message: fmt.Sprintf("failed to query source: %v", err),
		}
		return
	} else if kind != internal.SourceKindFile {
		outResp = vcrest.PatchFileSource409JSONResponse{
			Message: "source is not a file",
		}
		return
	}
	var body internal.FileSource
	if err := json.Unmarshal(rawBody, &body); err != nil {
		outResp = vcrest.PatchFileSource500JSONResponse{
			Message: fmt.Sprintf("failed to unmarshal source body: %v", err),
		}
		return
	}

	if updatePath {
		body.Path = newPath
	}

	rawBody, err = json.Marshal(body)
	if err != nil {
		outResp = vcrest.PatchFileSource500JSONResponse{
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
		outResp = vcrest.PatchFileSource500JSONResponse{
			Message: fmt.Sprintf("failed to update source: %v", err),
		}
		return
	}

	if err := txn.Commit(ctx); err != nil {
		outResp = vcrest.PatchFileSource500JSONResponse{
			Message: fmt.Sprintf("failed to commit transaction: %v", err),
		}
		return
	}

	outResp = vcrest.PatchFileSource200Response{}
	return
}
