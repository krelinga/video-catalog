package main

import (
	"context"
	"fmt"

	"github.com/krelinga/video-catalog/vcrest"
)

// PatchFileSource updates fields of a file source with the given UUID
func (s *Server) PatchFileSource(ctx context.Context, request vcrest.PatchFileSourceRequestObject) (vcrest.PatchFileSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}
