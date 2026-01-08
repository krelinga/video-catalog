package main

import (
	"context"
	"fmt"

	"github.com/krelinga/video-catalog/vcrest"
)

// ListPlans lists plans with optional filtering.
func (s *Server) ListPlans(ctx context.Context, request vcrest.ListPlansRequestObject) (vcrest.ListPlansResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}
