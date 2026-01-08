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
	requestUuid, err := internal.ParseUUID(request.Uuid.String())
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

	sourceUuid, updateInputs, err := internal.ValidateOptionalNonNullableUUID(request.Body.SourceUuid)
	if err != nil {
		outResp = vcrest.PatchChapterRangePlan400JSONResponse{
			Message: fmt.Sprintf("SourceUuid: %v", err),
		}
		return
	}

	workUuid, updateOutputs, err := internal.ValidateOptionalNonNullableUUID(request.Body.WorkUuid)
	if err != nil {
		outResp = vcrest.PatchChapterRangePlan400JSONResponse{
			Message: fmt.Sprintf("WorkUuid: %v", err),
		}
		return
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
	if updateInputs {
		body.SourceUUID = sourceUuid
	}
	if updateOutputs {
		body.WorkUUID = workUuid
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
		if err := internal.UpdatePlanInputs(ctx, txn, requestUuid, body.SourceUUID); err != nil {
			outResp = vcrest.PatchChapterRangePlan500JSONResponse{
				Message: err.Error(),
			}
			return
		}
	}

	if updateOutputs {
		if err := internal.UpdatePlanOutputs(ctx, txn, requestUuid, body.WorkUUID); err != nil {
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
