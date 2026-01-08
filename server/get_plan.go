package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
	"github.com/oapi-codegen/nullable"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// GetPlan retrieves a plan by UUID
func (s *Server) GetPlan(ctx context.Context, request vcrest.GetPlanRequestObject) (outResp vcrest.GetPlanResponseObject, outErr error) {
	// Validate request.
	requestUuid, err := uuid.Parse(request.Uuid.String())
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
			Uuid: request.Uuid,
			Direct: &vcrest.DirectPlan{
				SourceUuid: nullable.NewNullableWithValue(openapi_types.UUID(directBody.SourceUUID)),
				WorkUuid:   nullable.NewNullableWithValue(openapi_types.UUID(directBody.WorkUUID)),
			},
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
			Uuid: request.Uuid,
			ChapterRange: &vcrest.ChapterRangePlan{
				SourceUuid: nullable.NewNullableWithValue(openapi_types.UUID(chapterRangeBody.SourceUUID)),
				WorkUuid:   nullable.NewNullableWithValue(openapi_types.UUID(chapterRangeBody.WorkUUID)),
				StartChapter: func() nullable.Nullable[int32] {
					if chapterRangeBody.StartChapter != nil {
						return nullable.NewNullableWithValue(int32(*chapterRangeBody.StartChapter))
					}
					return nullable.Nullable[int32]{}
				}(),
				EndChapter: func() nullable.Nullable[int32] {
					if chapterRangeBody.EndChapter != nil {
						return nullable.NewNullableWithValue(int32(*chapterRangeBody.EndChapter))
					}
					return nullable.Nullable[int32]{}
				}(),
			},
		}
		return
	default:
		outResp = vcrest.GetPlan500JSONResponse{
			Message: fmt.Sprintf("unimplemented plan kind: %s", kind),
		}
		return
	}
}
