package main

import (
	"context"
	"fmt"

	"github.com/krelinga/video-catalog/vcrest"
)

// PatchMovieWork updates fields of a movie work with the given UUID
func (s *Server) PatchMovieWork(ctx context.Context, request vcrest.PatchMovieWorkRequestObject) (vcrest.PatchMovieWorkResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}