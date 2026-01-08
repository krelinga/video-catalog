package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
)

// PutFileSource adds or updates a file source with the given UUID
func (s *Server) PutFileSource(ctx context.Context, request vcrest.PutFileSourceRequestObject) (outResp vcrest.PutFileSourceResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.AsUUID(request.Uuid)
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
	if err := errors.Join(
		internal.FieldRequired(request.Body.Path),
		internal.FieldNotNull(request.Body.Path),
		internal.FieldNotEmpty(request.Body.Path),
	); err != nil {
		outResp = vcrest.PutFileSource400JSONResponse{
			Message: fmt.Sprintf("Path: %v", err),
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

	result, err := internal.UpsertEntity(ctx, s.Pool, "sources", requestUuid, internal.SourceKindFile, bodyRaw)
	if errors.Is(err, internal.ErrUpsertType) {
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

	if result == internal.UpsertCreated {
		outResp = vcrest.PutFileSource201Response{}
	} else {
		outResp = vcrest.PutFileSource200Response{}
	}
	return
}
