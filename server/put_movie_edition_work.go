package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
)

// PutMovieEdition creates or updates a movie edition work with the given UUID
func (s *Server) PutMovieEdition(ctx context.Context, request vcrest.PutMovieEditionRequestObject) (outResp vcrest.PutMovieEditionResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.AsUUID(request.Uuid)
	if err != nil {
		outResp = vcrest.PutMovieEdition400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PutMovieEdition400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	if err := errors.Join(
		internal.FieldRequired(request.Body.EditionType),
		internal.FieldNotNull(request.Body.EditionType),
		internal.FieldNotEmpty(request.Body.EditionType),
	); err != nil {
		outResp = vcrest.PutMovieEdition400JSONResponse{
			Message: fmt.Sprintf("EditionType: %v", err),
		}
		return
	}

	body := internal.MovieEditionWork{
		EditionType: request.Body.EditionType.MustGet(),
	}

	bodyRaw, err := json.Marshal(body)
	if err != nil {
		outResp = vcrest.PutMovieEdition500JSONResponse{
			Message: fmt.Sprintf("failed to marshal database body: %v", err),
		}
		return
	}

	result, err := internal.UpsertEntity(ctx, s.Pool, "works", requestUuid, internal.WorkKindMovieEdition, bodyRaw)
	if errors.Is(err, internal.ErrUpsertType) {
		outResp = vcrest.PutMovieEdition409JSONResponse{
			Message: "work with given UUID already exists with different kind",
		}
		return
	} else if err != nil {
		outResp = vcrest.PutMovieEdition500JSONResponse{
			Message: fmt.Sprintf("failed to insert/update work: %v", err),
		}
		return
	}

	if result == internal.UpsertCreated {
		outResp = vcrest.PutMovieEdition201Response{}
	} else {
		outResp = vcrest.PutMovieEdition200Response{}
	}
	return
}
