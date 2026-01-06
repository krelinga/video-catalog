package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
)

type Server struct {
	Config *internal.Config
	Pool *pgxpool.Pool
}

// GetSource retrieves a source by UUID
func (s *Server) GetSource(ctx context.Context, request vcrest.GetSourceRequestObject) (vcrest.GetSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// PutDiscSource adds or updates a disc source with the given UUID
func (s *Server) PutDiscSource(ctx context.Context, request vcrest.PutDiscSourceRequestObject) (vcrest.PutDiscSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// PutFileSource adds or updates a file source with the given UUID
func (s *Server) PutFileSource(ctx context.Context, request vcrest.PutFileSourceRequestObject) (vcrest.PutFileSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// PutMovieEdition creates a new movie edition work with the given UUID
func (s *Server) PutMovieEdition(ctx context.Context, request vcrest.PutMovieEditionRequestObject) (vcrest.PutMovieEditionResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) PatchMovieEdition(ctx context.Context, request vcrest.PatchMovieEditionRequestObject) (vcrest.PatchMovieEditionResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}
