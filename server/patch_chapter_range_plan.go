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

// PatchChapterRangePlan updates fields of a chapter range plan with the given UUID
func (s *Server) PatchChapterRangePlan(ctx context.Context, request vcrest.PatchChapterRangePlanRequestObject) (outResp vcrest.PatchChapterRangePlanResponseObject, _ error) {
	// Validate request.
	requestUuid, err := uuid.Parse(request.Uuid.String())
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

	var sourceUuid uuid.UUID
	if request.Body.SourceUuid.IsSpecified() {
		if request.Body.SourceUuid.IsNull() {
			outResp = vcrest.PatchChapterRangePlan400JSONResponse{
				Message: "SourceUuid cannot be null",
			}
			return
		}
		sourceUuid, err = uuid.Parse(request.Body.SourceUuid.MustGet().String())
		if err != nil {
			outResp = vcrest.PatchChapterRangePlan400JSONResponse{
				Message: "invalid SourceUuid format",
			}
			return
		}
		if sourceUuid == uuid.Nil {
			outResp = vcrest.PatchChapterRangePlan400JSONResponse{
				Message: "SourceUuid cannot be empty",
			}
			return
		}
	}

	var workUuid uuid.UUID
	if request.Body.WorkUuid.IsSpecified() {
		if request.Body.WorkUuid.IsNull() {
			outResp = vcrest.PatchChapterRangePlan400JSONResponse{
				Message: "WorkUuid cannot be null",
			}
			return
		}
		workUuid, err = uuid.Parse(request.Body.WorkUuid.MustGet().String())
		if err != nil {
			outResp = vcrest.PatchChapterRangePlan400JSONResponse{
				Message: "invalid WorkUuid format",
			}
			return
		}
		if workUuid == uuid.Nil {
			outResp = vcrest.PatchChapterRangePlan400JSONResponse{
				Message: "WorkUuid cannot be empty",
			}
			return
		}
	}

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

	// Track if we need to update the relation tables
	updateInputs := false
	updateOutputs := false

	if request.Body.SourceUuid.IsSpecified() {
		body.SourceUUID = sourceUuid
		updateInputs = true
	}
	if request.Body.WorkUuid.IsSpecified() {
		body.WorkUUID = workUuid
		updateOutputs = true
	}
	if request.Body.StartChapter.IsSpecified() {
		if request.Body.StartChapter.IsNull() {
			body.StartChapter = nil
		} else {
			startChap := int(request.Body.StartChapter.MustGet())
			body.StartChapter = &startChap
		}
	}
	if request.Body.EndChapter.IsSpecified() {
		if request.Body.EndChapter.IsNull() {
			body.EndChapter = nil
		} else {
			endChap := int(request.Body.EndChapter.MustGet())
			body.EndChapter = &endChap
		}
	}

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

	if updateInputs {
		_, err = txn.Exec(ctx, `
			DELETE FROM plan_inputs WHERE plan_uuid = $1
		`, requestUuid)
		if err != nil {
			outResp = vcrest.PatchChapterRangePlan500JSONResponse{
				Message: fmt.Sprintf("failed to delete old plan_inputs: %v", err),
			}
			return
		}

		_, err = txn.Exec(ctx, `
			INSERT INTO plan_inputs (plan_uuid, source_uuid)
			VALUES ($1, $2)
		`, requestUuid, body.SourceUUID)
		if err != nil {
			outResp = vcrest.PatchChapterRangePlan500JSONResponse{
				Message: fmt.Sprintf("failed to insert plan_inputs: %v", err),
			}
			return
		}
	}

	if updateOutputs {
		_, err = txn.Exec(ctx, `
			DELETE FROM plan_outputs WHERE plan_uuid = $1
		`, requestUuid)
		if err != nil {
			outResp = vcrest.PatchChapterRangePlan500JSONResponse{
				Message: fmt.Sprintf("failed to delete old plan_outputs: %v", err),
			}
			return
		}

		_, err = txn.Exec(ctx, `
			INSERT INTO plan_outputs (plan_uuid, work_uuid)
			VALUES ($1, $2)
		`, requestUuid, body.WorkUUID)
		if err != nil {
			outResp = vcrest.PatchChapterRangePlan500JSONResponse{
				Message: fmt.Sprintf("failed to insert plan_outputs: %v", err),
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
