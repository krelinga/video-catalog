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

// PatchDirectPlan updates fields of a direct plan with the given UUID
func (s *Server) PatchDirectPlan(ctx context.Context, request vcrest.PatchDirectPlanRequestObject) (outResp vcrest.PatchDirectPlanResponseObject, _ error) {
	// Validate request.
	requestUuid, err := uuid.Parse(request.Uuid.String())
	if err != nil {
		outResp = vcrest.PatchDirectPlan400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PatchDirectPlan400JSONResponse{
			Message: "request body is required",
		}
		return
	}

	var sourceUuid uuid.UUID
	if request.Body.SourceUuid.IsSpecified() {
		if request.Body.SourceUuid.IsNull() {
			outResp = vcrest.PatchDirectPlan400JSONResponse{
				Message: "SourceUuid cannot be null",
			}
			return
		}
		sourceUuid, err = uuid.Parse(request.Body.SourceUuid.MustGet().String())
		if err != nil {
			outResp = vcrest.PatchDirectPlan400JSONResponse{
				Message: "invalid SourceUuid format",
			}
			return
		}
		if sourceUuid == uuid.Nil {
			outResp = vcrest.PatchDirectPlan400JSONResponse{
				Message: "SourceUuid cannot be empty",
			}
			return
		}
	}

	var workUuid uuid.UUID
	if request.Body.WorkUuid.IsSpecified() {
		if request.Body.WorkUuid.IsNull() {
			outResp = vcrest.PatchDirectPlan400JSONResponse{
				Message: "WorkUuid cannot be null",
			}
			return
		}
		workUuid, err = uuid.Parse(request.Body.WorkUuid.MustGet().String())
		if err != nil {
			outResp = vcrest.PatchDirectPlan400JSONResponse{
				Message: "invalid WorkUuid format",
			}
			return
		}
		if workUuid == uuid.Nil {
			outResp = vcrest.PatchDirectPlan400JSONResponse{
				Message: "WorkUuid cannot be empty",
			}
			return
		}
	}

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.PatchDirectPlan500JSONResponse{
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
		outResp = vcrest.PatchDirectPlan404JSONResponse{
			Message: "plan not found",
		}
		return
	} else if err != nil {
		outResp = vcrest.PatchDirectPlan500JSONResponse{
			Message: fmt.Sprintf("failed to query plan: %v", err),
		}
		return
	} else if kind != internal.PlanKindDirect {
		outResp = vcrest.PatchDirectPlan409JSONResponse{
			Message: "plan is not a direct plan",
		}
		return
	}

	var body internal.DirectPlan
	if err := json.Unmarshal(rawBody, &body); err != nil {
		outResp = vcrest.PatchDirectPlan500JSONResponse{
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

	rawBody, err = json.Marshal(body)
	if err != nil {
		outResp = vcrest.PatchDirectPlan500JSONResponse{
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
		outResp = vcrest.PatchDirectPlan500JSONResponse{
			Message: fmt.Sprintf("failed to update plan: %v", err),
		}
		return
	}

	if updateInputs {
		_, err = txn.Exec(ctx, `
			DELETE FROM plan_inputs WHERE plan_uuid = $1
		`, requestUuid)
		if err != nil {
			outResp = vcrest.PatchDirectPlan500JSONResponse{
				Message: fmt.Sprintf("failed to delete old plan_inputs: %v", err),
			}
			return
		}

		_, err = txn.Exec(ctx, `
			INSERT INTO plan_inputs (plan_uuid, source_uuid)
			VALUES ($1, $2)
		`, requestUuid, body.SourceUUID)
		if err != nil {
			outResp = vcrest.PatchDirectPlan500JSONResponse{
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
			outResp = vcrest.PatchDirectPlan500JSONResponse{
				Message: fmt.Sprintf("failed to delete old plan_outputs: %v", err),
			}
			return
		}

		_, err = txn.Exec(ctx, `
			INSERT INTO plan_outputs (plan_uuid, work_uuid)
			VALUES ($1, $2)
		`, requestUuid, body.WorkUUID)
		if err != nil {
			outResp = vcrest.PatchDirectPlan500JSONResponse{
				Message: fmt.Sprintf("failed to insert plan_outputs: %v", err),
			}
			return
		}
	}

	if err := txn.Commit(ctx); err != nil {
		outResp = vcrest.PatchDirectPlan500JSONResponse{
			Message: fmt.Sprintf("failed to commit transaction: %v", err),
		}
		return
	}

	outResp = vcrest.PatchDirectPlan200Response{}
	return
}
