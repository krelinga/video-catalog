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
)

// PutChapterRangePlan adds or updates a chapter range plan with the given UUID
func (s *Server) PutChapterRangePlan(ctx context.Context, request vcrest.PutChapterRangePlanRequestObject) (outResp vcrest.PutChapterRangePlanResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.AsUUID(request.Uuid)
	if err != nil {
		outResp = vcrest.PutChapterRangePlan400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if requestUuid == uuid.Nil {
		outResp = vcrest.PutChapterRangePlan400JSONResponse{
			Message: "UUID cannot be zero",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PutChapterRangePlan400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	if err := errors.Join(
		internal.FieldRequired(request.Body.SourceUuid),
		internal.FieldNotNull(request.Body.SourceUuid),
		internal.FieldValidUUID(request.Body.SourceUuid),
	); err != nil {
		outResp = vcrest.PutChapterRangePlan400JSONResponse{
			Message: fmt.Sprintf("SourceUuid: %v", err),
		}
		return
	}
	sourceUuid := internal.FieldMustUUID(request.Body.SourceUuid)

	if err := errors.Join(
		internal.FieldRequired(request.Body.WorkUuid),
		internal.FieldNotNull(request.Body.WorkUuid),
		internal.FieldValidUUID(request.Body.WorkUuid),
	); err != nil {
		outResp = vcrest.PutChapterRangePlan400JSONResponse{
			Message: fmt.Sprintf("WorkUuid: %v", err),
		}
		return
	}
	workUuid := internal.FieldMustUUID(request.Body.WorkUuid)

	body := internal.ChapterRangePlan{
		SourceUUID: sourceUuid,
		WorkUUID:   workUuid,
	}
	internal.FieldSet(request.Body.StartChapter, body.StartChapter)
	internal.FieldSet(request.Body.EndChapter, body.EndChapter)

	bodyRaw, err := json.Marshal(body)
	if err != nil {
		outResp = vcrest.PutChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to marshal database body: %v", err),
		}
		return
	}

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.PutChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to begin transaction: %v", err),
		}
		return
	}
	defer txn.Rollback(ctx)

	row := txn.QueryRow(ctx, `
		INSERT INTO plans (uuid, kind, body)
		VALUES ($1, $2, $3)
		ON CONFLICT (uuid) DO UPDATE
		SET body = EXCLUDED.body
		WHERE plans.kind = $2
		RETURNING xmax`,
		requestUuid, internal.PlanKindChapterRange, bodyRaw)
	var xmax uint32
	if err := row.Scan(&xmax); errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.PutChapterRangePlan409JSONResponse{
			Message: "plan with given UUID already exists with different kind",
		}
		return
	} else if err != nil {
		outResp = vcrest.PutChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to insert/update plan: %v", err),
		}
		return
	}

	// Update plan_inputs
	if err := internal.UpdatePlanInputs(ctx, txn, requestUuid, sourceUuid); err != nil {
		outResp = vcrest.PutChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to update plan_inputs: %v", err),
		}
		return
	}

	// Update plan_outputs
	if err := internal.UpdatePlanOutputs(ctx, txn, requestUuid, workUuid); err != nil {
		outResp = vcrest.PutChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to update plan_outputs: %v", err),
		}
		return
	}

	if err := txn.Commit(ctx); err != nil {
		outResp = vcrest.PutChapterRangePlan500JSONResponse{
			Message: fmt.Sprintf("failed to commit transaction: %v", err),
		}
		return
	}

	if xmax == 0 {
		outResp = vcrest.PutChapterRangePlan201Response{}
	} else {
		outResp = vcrest.PutChapterRangePlan200Response{}
	}
	return
}
