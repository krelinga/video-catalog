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

// PatchDiscSource updates fields of a disc source with the given UUID
func (s *Server) PatchDiscSource(ctx context.Context, request vcrest.PatchDiscSourceRequestObject) (outResp vcrest.PatchDiscSourceResponseObject, _ error) {
	// Validate request.
	requestUuid, err := uuid.Parse(request.Uuid.String())
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
	if request.Body.OrigDirName.IsSpecified() && (request.Body.OrigDirName.IsNull() || request.Body.OrigDirName.MustGet() == "") {
		outResp = vcrest.PatchDiscSource400JSONResponse{
			Message: "OrigDirName cannot be null or empty",
		}
		return
	}
	if request.Body.Path.IsSpecified() && (request.Body.Path.IsNull() || request.Body.Path.MustGet() == "") {
		outResp = vcrest.PatchDiscSource400JSONResponse{
			Message: "Path cannot be null or empty",
		}
		return
	}
	if request.Body.AllFilesAdded.IsSpecified() && request.Body.AllFilesAdded.IsNull() {
		outResp = vcrest.PatchDiscSource400JSONResponse{
			Message: "AllFilesAdded cannot be null",
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

	if request.Body.OrigDirName.IsSpecified() {
		body.OrigDirName = request.Body.OrigDirName.MustGet()
	}
	if request.Body.Path.IsSpecified() {
		body.Path = request.Body.Path.MustGet()
	}
	if request.Body.AllFilesAdded.IsSpecified() {
		body.AllFilesAdded = request.Body.AllFilesAdded.MustGet()
	}

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
