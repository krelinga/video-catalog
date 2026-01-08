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
	Pool   *pgxpool.Pool
}

// GetPlan retrieves a plan by UUID
func (s *Server) GetPlan(ctx context.Context, request vcrest.GetPlanRequestObject) (vcrest.GetPlanResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// PutDirectPlan adds or updates a direct plan with the given UUID
func (s *Server) PutDirectPlan(ctx context.Context, request vcrest.PutDirectPlanRequestObject) (vcrest.PutDirectPlanResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// PatchDirectPlan updates fields of a direct plan with the given UUID
func (s *Server) PatchDirectPlan(ctx context.Context, request vcrest.PatchDirectPlanRequestObject) (vcrest.PatchDirectPlanResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// PutChapterRangePlan adds or updates a chapter range plan with the given UUID
func (s *Server) PutChapterRangePlan(ctx context.Context, request vcrest.PutChapterRangePlanRequestObject) (vcrest.PutChapterRangePlanResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// PatchChapterRangePlan updates fields of a chapter range plan with the given UUID
func (s *Server) PatchChapterRangePlan(ctx context.Context, request vcrest.PatchChapterRangePlanRequestObject) (vcrest.PatchChapterRangePlanResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// ListPlans lists plans with optional filtering.
func (s *Server) ListPlans(ctx context.Context, request vcrest.ListPlansRequestObject) (vcrest.ListPlansResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}
