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
	requestUuid, err := internal.ParseUUID(request.Uuid.String())
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
	// TODO: call helper methods.
	var body internal.MovieWork
	if !request.Body.Title.IsSpecified() || request.Body.Title.IsNull() || request.Body.Title.MustGet() == "" {
		outResp = vcrest.PutMovieWork400JSONResponse{
			Message: "non-empty title is required",
		}
		return
	} else {
		body.Title = request.Body.Title.MustGet()
	}
	if request.Body.ReleaseYear.IsSpecified() && !request.Body.ReleaseYear.IsNull() {
		ry := request.Body.ReleaseYear.MustGet()
		body.ReleaseYear = &ry
	}
	if request.Body.TmdbId.IsSpecified() && !request.Body.TmdbId.IsNull() {
		tid := request.Body.TmdbId.MustGet()
		body.TmdbId = &tid
	}

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
