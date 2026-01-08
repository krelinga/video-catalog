package main

import (
	"context"
	"fmt"

	"github.com/krelinga/video-catalog/vcrest"
)

// PatchDirectPlan updates fields of a direct plan with the given UUID
func (s *Server) PatchDirectPlan(ctx context.Context, request vcrest.PatchDirectPlanRequestObject) (vcrest.PatchDirectPlanResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}
