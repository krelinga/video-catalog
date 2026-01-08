package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
)

// PutDiscSource adds or updates a disc source with the given UUID
func (s *Server) PutDiscSource(ctx context.Context, request vcrest.PutDiscSourceRequestObject) (outResp vcrest.PutDiscSourceResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.AsUUID(request.Uuid)
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
	if err := errors.Join(
		internal.FieldRequired(request.Body.OrigDirName),
		internal.FieldNotNull(request.Body.OrigDirName),
		internal.FieldNotEmpty(request.Body.OrigDirName),
	); err != nil {
		outResp = vcrest.PutDiscSource400JSONResponse{
			Message: fmt.Sprintf("OrigDirName: %v", err),
		}
		return
	}

	if err := errors.Join(
		internal.FieldRequired(request.Body.Path),
		internal.FieldNotNull(request.Body.Path),
		internal.FieldNotEmpty(request.Body.Path),
	); err != nil {
		outResp = vcrest.PutDiscSource400JSONResponse{
			Message: fmt.Sprintf("Path: %v", err),
		}
		return
	}

	if err := internal.FieldNotNull(request.Body.AllFilesAdded); err != nil {
		outResp = vcrest.PutDiscSource400JSONResponse{
			Message: fmt.Sprintf("AllFilesAdded: %v", err),
		}
		return
	}

	body := internal.DiscSource{
		OrigDirName: request.Body.OrigDirName.MustGet(),
		Path:        request.Body.Path.MustGet(),
	}
	internal.FieldSet(request.Body.AllFilesAdded, &body.AllFilesAdded)

	bodyRaw, err := json.Marshal(body)
	if err != nil {
		outResp = vcrest.PutDiscSource500JSONResponse{
			Message: fmt.Sprintf("failed to marshal database body: %v", err),
		}
		return
	}

	result, err := internal.UpsertEntity(ctx, s.Pool, "sources", requestUuid, internal.SourceKindDisc, bodyRaw)
	if errors.Is(err, internal.ErrUpsertType) {
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

	if result == internal.UpsertCreated {
		outResp = vcrest.PutDiscSource201Response{}
	} else {
		outResp = vcrest.PutDiscSource200Response{}
	}
	return
}
