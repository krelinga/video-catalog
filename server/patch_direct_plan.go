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

// PatchDirectPlan updates fields of a direct plan with the given UUID
func (s *Server) PatchDirectPlan(ctx context.Context, request vcrest.PatchDirectPlanRequestObject) (outResp vcrest.PatchDirectPlanResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.ParseUUID(request.Uuid.String())
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

	sourceUuid, updateInputs, err := internal.ValidateOptionalNonNullableUUID(request.Body.SourceUuid)
	if err != nil {
		outResp = vcrest.PatchDirectPlan400JSONResponse{
			Message: fmt.Sprintf("SourceUuid: %v", err),
		}
		return
	}

	workUuid, updateOutputs, err := internal.ValidateOptionalNonNullableUUID(request.Body.WorkUuid)
	if err != nil {
		outResp = vcrest.PatchDirectPlan400JSONResponse{
			Message: fmt.Sprintf("WorkUuid: %v", err),
		}
		return
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
	if updateInputs {
		body.SourceUUID = sourceUuid
	}
	if updateOutputs {
		body.WorkUUID = workUuid
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
		if err := internal.UpdatePlanInputs(ctx, txn, requestUuid, body.SourceUUID); err != nil {
			outResp = vcrest.PatchDirectPlan500JSONResponse{
				Message: err.Error(),
			}
			return
		}
	}

	if updateOutputs {
		if err := internal.UpdatePlanOutputs(ctx, txn, requestUuid, body.WorkUUID); err != nil {
			outResp = vcrest.PatchDirectPlan500JSONResponse{
				Message: err.Error(),
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
