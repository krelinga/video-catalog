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

// PatchChapterRangePlan updates fields of a chapter range plan with the given UUID
func (s *Server) PatchChapterRangePlan(ctx context.Context, request vcrest.PatchChapterRangePlanRequestObject) (outResp vcrest.PatchChapterRangePlanResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.AsUUID(request.Uuid)
	if err != nil {
		outResp = vcrest.PatchChapterRangePlan400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PatchChapterRangePlan400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	if err := errors.Join(
		internal.FieldNotNull(request.Body.SourceUuid),
		internal.FieldValidUUID(request.Body.SourceUuid),
	); err != nil {
		outResp = vcrest.PatchChapterRangePlan400JSONResponse{
			Message: fmt.Sprintf("SourceUuid: %v", err),
		}
		return
	}
	sourceUuid := internal.FieldMayUUID(request.Body.SourceUuid)

	if err := errors.Join(
		internal.FieldNotNull(request.Body.WorkUuid),
		internal.FieldValidUUID(request.Body.WorkUuid),
	); err != nil {
		outResp = vcrest.PatchChapterRangePlan400JSONResponse{
			Message: fmt.Sprintf("WorkUuid: %v", err),
		}
		return
	}
	workUuid := internal.FieldMayUUID(request.Body.WorkUuid)

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.PatchChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to begin transaction: %v", err),
		}
		return
	}
	defer txn.Rollback(ctx)

	var kind internal.PlanKind
	var rawBody json.RawMessage
	row := txn.QueryRow(ctx, `
		SELECT kind, body
		FROM plans
		WHERE uuid = $1
	`, requestUuid)
	err = row.Scan(&kind, &rawBody)
	if errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.PatchChapterRangePlan404JSONResponse{
			Message: "plan not found",
		}
		return
	} else if err != nil {
		outResp = vcrest.PatchChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to query plan: %v", err),
		}
		return
	} else if kind != internal.PlanKindChapterRange {
		outResp = vcrest.PatchChapterRangePlan409JSONResponse{
			Message: "plan is not a chapter range plan",
		}
		return
	}

	var body internal.ChapterRangePlan
	if err := json.Unmarshal(rawBody, &body); err != nil {
		outResp = vcrest.PatchChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to unmarshal plan body: %v", err),
		}
		return
	}

	if sourceUuid != nil {
		body.SourceUUID = *sourceUuid
	}
	if workUuid != nil {
		body.WorkUUID = *workUuid
	}
	internal.FieldSetClear(request.Body.StartChapter, &body.StartChapter)
	internal.FieldSetClear(request.Body.EndChapter, &body.EndChapter)

	rawBody, err = json.Marshal(body)
	if err != nil {
		outResp = vcrest.PatchChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to marshal database body: %v", err),
		}
		return
	}

	_, err = txn.Exec(ctx, `
		UPDATE plans
		SET body = $2
		WHERE uuid = $1
	`, requestUuid, rawBody)
	if err != nil {
		outResp = vcrest.PatchChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to update plan: %v", err),
		}
		return
	}

	if sourceUuid != nil {
		if err := internal.UpdatePlanInputs(ctx, txn, requestUuid, *sourceUuid); err != nil {
			outResp = vcrest.PatchChapterRangePlan500JSONResponse{
				Message: err.Error(),
			}
			return
		}
	}

	if workUuid != nil {
		if err := internal.UpdatePlanOutputs(ctx, txn, requestUuid, *workUuid); err != nil {
			outResp = vcrest.PatchChapterRangePlan500JSONResponse{
				Message: err.Error(),
			}
			return
		}
	}

	if err := txn.Commit(ctx); err != nil {
		outResp = vcrest.PatchChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to commit transaction: %v", err),
		}
		return
	}

	outResp = vcrest.PatchChapterRangePlan200Response{}
	return
}
