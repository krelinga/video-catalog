package main

import (
	"context"
	"fmt"

	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
)

type Server struct {
	Config *internal.Config
}

// AddFileToDisc associates a file source with a disc source
func (s *Server) AddFileToDisc(ctx context.Context, request vcrest.AddFileToDiscRequestObject) (vcrest.AddFileToDiscResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetSource retrieves a source by UUID
func (s *Server) GetSource(ctx context.Context, request vcrest.GetSourceRequestObject) (vcrest.GetSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// AddDiscSource adds a disc source with the given UUID
func (s *Server) AddDiscSource(ctx context.Context, request vcrest.AddDiscSourceRequestObject) (vcrest.AddDiscSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// MarkDiscAllFilesAdded marks all files from the disc source as added
func (s *Server) MarkDiscAllFilesAdded(ctx context.Context, request vcrest.MarkDiscAllFilesAddedRequestObject) (vcrest.MarkDiscAllFilesAddedResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// AddFileSource adds a file source with the given UUID
func (s *Server) AddFileSource(ctx context.Context, request vcrest.AddFileSourceRequestObject) (vcrest.AddFileSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// AssociateMovieEdition associates an existing movie edition with a movie work
func (s *Server) AssociateMovieEdition(ctx context.Context, request vcrest.AssociateMovieEditionRequestObject) (vcrest.AssociateMovieEditionResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetWork retrieves a work by UUID
func (s *Server) GetWork(ctx context.Context, request vcrest.GetWorkRequestObject) (vcrest.GetWorkResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// AddMovieWork adds a movie work with the given UUID
func (s *Server) AddMovieWork(ctx context.Context, request vcrest.AddMovieWorkRequestObject) (vcrest.AddMovieWorkResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// AddMovieEdition creates a new movie edition work with the given UUID
func (s *Server) AddMovieEdition(ctx context.Context, request vcrest.AddMovieEditionRequestObject) (vcrest.AddMovieEditionResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}
