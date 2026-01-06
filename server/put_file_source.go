package main

import (
	"context"
	"fmt"

	"github.com/krelinga/video-catalog/vcrest"
)

// PutFileSource adds or updates a file source with the given UUID
func (s *Server) PutFileSource(ctx context.Context, request vcrest.PutFileSourceRequestObject) (vcrest.PutFileSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}