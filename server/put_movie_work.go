package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
)

// PutMovieWork adds or updates a movie work with the given UUID
func (s *Server) PutMovieWork(ctx context.Context, request vcrest.PutMovieWorkRequestObject) (outResp vcrest.PutMovieWorkResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.AsUUID(request.Uuid)
	if err != nil {
		outResp = vcrest.PutMovieWork400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PutMovieWork400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	if err := errors.Join(
		internal.FieldRequired(request.Body.Title),
		internal.FieldNotNull(request.Body.Title),
		internal.FieldNotEmpty(request.Body.Title),
	); err != nil {
		outResp = vcrest.PutMovieWork400JSONResponse{
			Message: fmt.Sprintf("Title: %v", err),
		}
		return
	}
	body := internal.MovieWork{
		Title: request.Body.Title.MustGet(),
	}
	internal.FieldSet(request.Body.ReleaseYear, body.ReleaseYear)
	internal.FieldSet(request.Body.TmdbId, body.TmdbId)
	
	bodyRaw, err := json.Marshal(body)
	if err != nil {
		outResp = vcrest.PutMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to marshal database body: %v", err),
		}
		return
	}

	result, err := internal.UpsertEntity(ctx, s.Pool, "works", requestUuid, internal.WorkKindMovie, bodyRaw)
	if errors.Is(err, internal.ErrUpsertType) {
		outResp = vcrest.PutMovieWork409JSONResponse{
			Message: "work with given UUID already exists with different kind",
		}
		return
	} else if err != nil {
		outResp = vcrest.PutMovieWork500JSONResponse{
			Message: fmt.Sprintf("failed to insert/update work: %v", err),
		}
		return
	}

	if result == internal.UpsertCreated {
		outResp = vcrest.PutMovieWork201Response{}
	} else {
		outResp = vcrest.PutMovieWork200Response{}
	}
	return
}
