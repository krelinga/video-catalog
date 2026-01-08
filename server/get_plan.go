package main

import (
	"context"
	"fmt"

	"github.com/krelinga/video-catalog/vcrest"
)

// GetPlan retrieves a plan by UUID
func (s *Server) GetPlan(ctx context.Context, request vcrest.GetPlanRequestObject) (vcrest.GetPlanResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}
