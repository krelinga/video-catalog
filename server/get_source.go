package main

import (
	"context"
	"fmt"

	"github.com/krelinga/video-catalog/vcrest"
)

// GetSource retrieves a source by UUID
func (s *Server) GetSource(ctx context.Context, request vcrest.GetSourceRequestObject) (vcrest.GetSourceResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}