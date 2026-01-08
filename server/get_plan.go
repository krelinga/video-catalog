package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
)

// GetPlan retrieves a plan by UUID
func (s *Server) GetPlan(ctx context.Context, request vcrest.GetPlanRequestObject) (outResp vcrest.GetPlanResponseObject, outErr error) {
	// Validate request.
	requestUuid, err := internal.ParseUUID(request.Uuid.String())
	if err != nil {
		outResp = vcrest.GetPlan400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.GetPlan500JSONResponse{
			Message: fmt.Sprintf("failed to begin transaction: %v", err),
		}
		return
	}
	defer txn.Rollback(ctx)

	var kind internal.PlanKind
	var bodyRaw json.RawMessage
	err = txn.QueryRow(ctx, `
		SELECT kind, body
		FROM plans
		WHERE uuid = $1
	`, requestUuid).Scan(&kind, &bodyRaw)
	if errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.GetPlan404JSONResponse{
			Message: "plan not found",
		}
		return
	} else if err != nil {
		outResp = vcrest.GetPlan500JSONResponse{
			Message: fmt.Sprintf("failed to query plan: %v", err),
		}
		return
	}

	if !kind.IsValid() {
		outResp = vcrest.GetPlan500JSONResponse{
			Message: fmt.Sprintf("invalid plan kind in database: %s", kind),
		}
		return
	}

	switch kind {
	case internal.PlanKindDirect:
		var directBody internal.DirectPlan
		if err := json.Unmarshal(bodyRaw, &directBody); err != nil {
			outResp = vcrest.GetPlan500JSONResponse{
				Message: fmt.Sprintf("failed to unmarshal direct plan body: %v", err),
			}
			return
		}
		outResp = vcrest.GetPlan200JSONResponse{
			Uuid:   request.Uuid,
			Direct: directBody.ToAPI(),
		}
		return
	case internal.PlanKindChapterRange:
		var chapterRangeBody internal.ChapterRangePlan
		if err := json.Unmarshal(bodyRaw, &chapterRangeBody); err != nil {
			outResp = vcrest.GetPlan500JSONResponse{
				Message: fmt.Sprintf("failed to unmarshal chapter range plan body: %v", err),
			}
			return
		}
		outResp = vcrest.GetPlan200JSONResponse{
			Uuid:         request.Uuid,
			ChapterRange: chapterRangeBody.ToAPI(),
		}
		return
	default:
		outResp = vcrest.GetPlan500JSONResponse{
			Message: fmt.Sprintf("unimplemented plan kind: %s", kind),
		}
		return
	}
}
