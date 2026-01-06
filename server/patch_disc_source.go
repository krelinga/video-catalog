package main

import (
	"context"
	"fmt"

	"github.com/krelinga/video-catalog/vcrest"
)

// PatchDiscSource updates fields of a disc source with the given UUID
func (s *Server) PatchDiscSource(ctx context.Context, request vcrest.PatchDiscSourceRequestObject) (vcrest.PatchDiscSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}