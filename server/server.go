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

// AddFileToDisc associates a file source with a disc source
func (s *Server) AddFileToDisc(ctx context.Context, request vcrest.AddFileToDiscRequestObject) (vcrest.AddFileToDiscResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetSource retrieves a source by UUID
func (s *Server) GetSource(ctx context.Context, request vcrest.GetSourceRequestObject) (vcrest.GetSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// PutDiscSource adds or updates a disc source with the given UUID
func (s *Server) PutDiscSource(ctx context.Context, request vcrest.PutDiscSourceRequestObject) (vcrest.PutDiscSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// MarkDiscAllFilesAdded marks all files from the disc source as added
func (s *Server) MarkDiscAllFilesAdded(ctx context.Context, request vcrest.MarkDiscAllFilesAddedRequestObject) (vcrest.MarkDiscAllFilesAddedResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// PutFileSource adds or updates a file source with the given UUID
func (s *Server) PutFileSource(ctx context.Context, request vcrest.PutFileSourceRequestObject) (vcrest.PutFileSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// AddEditionToMovie associates an existing movie edition with a movie work
func (s *Server) AddEditionToMovie(ctx context.Context, request vcrest.AddEditionToMovieRequestObject) (vcrest.AddEditionToMovieResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetWork retrieves a work by UUID
func (s *Server) GetWork(ctx context.Context, request vcrest.GetWorkRequestObject) (vcrest.GetWorkResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// PutMovieWork adds or updates a movie work with the given UUID
func (s *Server) PutMovieWork(ctx context.Context, request vcrest.PutMovieWorkRequestObject) (vcrest.PutMovieWorkResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// PutMovieEdition creates a new movie edition work with the given UUID
func (s *Server) PutMovieEdition(ctx context.Context, request vcrest.PutMovieEditionRequestObject) (vcrest.PutMovieEditionResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}
