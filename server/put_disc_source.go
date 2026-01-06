package main

import (
	"context"
	"fmt"

	"github.com/krelinga/video-catalog/vcrest"
)

// PutDiscSource adds or updates a disc source with the given UUID
func (s *Server) PutDiscSource(ctx context.Context, request vcrest.PutDiscSourceRequestObject) (vcrest.PutDiscSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}