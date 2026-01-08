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

// PutDirectPlan adds or updates a direct plan with the given UUID
func (s *Server) PutDirectPlan(ctx context.Context, request vcrest.PutDirectPlanRequestObject) (outResp vcrest.PutDirectPlanResponseObject, _ error) {
	// Validate request.
	requestUuid, err := uuid.Parse(request.Uuid.String())
	if err != nil {
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	if !request.Body.SourceUuid.IsSpecified() || request.Body.SourceUuid.IsNull() {
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: "SourceUuid is required",
		}
		return
	}
	sourceUuid, err := uuid.Parse(request.Body.SourceUuid.MustGet().String())
	if err != nil {
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: "invalid SourceUuid format",
		}
		return
	}
	if sourceUuid == uuid.Nil {
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: "SourceUuid cannot be empty",
		}
		return
	}
	if !request.Body.WorkUuid.IsSpecified() || request.Body.WorkUuid.IsNull() {
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: "WorkUuid is required",
		}
		return
	}
	workUuid, err := uuid.Parse(request.Body.WorkUuid.MustGet().String())
	if err != nil {
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: "invalid WorkUuid format",
		}
		return
	}
	if workUuid == uuid.Nil {
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: "WorkUuid cannot be empty",
		}
		return
	}

	body := internal.DirectPlan{
		SourceUUID: sourceUuid,
		WorkUUID:   workUuid,
	}

	bodyRaw, err := json.Marshal(body)
	if err != nil {
		outResp = vcrest.PutDirectPlan500JSONResponse{
			Message: fmt.Sprintf("failed to marshal database body: %v", err),
		}
		return
	}

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.PutDirectPlan500JSONResponse{
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
		requestUuid, internal.PlanKindDirect, bodyRaw)
	var xmax uint32
	if err := row.Scan(&xmax); errors.Is(err, pgx.ErrNoRows) {
		outResp = vcrest.PutDirectPlan409JSONResponse{
			Message: "plan with given UUID already exists with different kind",
		}
		return
	} else if err != nil {
		outResp = vcrest.PutDirectPlan500JSONResponse{
			Message: fmt.Sprintf("failed to insert/update plan: %v", err),
		}
		return
	}

	// Update plan_inputs
	_, err = txn.Exec(ctx, `
		DELETE FROM plan_inputs WHERE plan_uuid = $1
	`, requestUuid)
	if err != nil {
		outResp = vcrest.PutDirectPlan500JSONResponse{
			Message: fmt.Sprintf("failed to delete old plan_inputs: %v", err),
		}
		return
	}

	_, err = txn.Exec(ctx, `
		INSERT INTO plan_inputs (plan_uuid, source_uuid)
		VALUES ($1, $2)
	`, requestUuid, sourceUuid)
	if err != nil {
		outResp = vcrest.PutDirectPlan500JSONResponse{
			Message: fmt.Sprintf("failed to insert plan_inputs: %v", err),
		}
		return
	}

	// Update plan_outputs
	_, err = txn.Exec(ctx, `
		DELETE FROM plan_outputs WHERE plan_uuid = $1
	`, requestUuid)
	if err != nil {
		outResp = vcrest.PutDirectPlan500JSONResponse{
			Message: fmt.Sprintf("failed to delete old plan_outputs: %v", err),
		}
		return
	}

	_, err = txn.Exec(ctx, `
		INSERT INTO plan_outputs (plan_uuid, work_uuid)
		VALUES ($1, $2)
	`, requestUuid, workUuid)
	if err != nil {
		outResp = vcrest.PutDirectPlan500JSONResponse{
			Message: fmt.Sprintf("failed to insert plan_outputs: %v", err),
		}
		return
	}

	if err := txn.Commit(ctx); err != nil {
		outResp = vcrest.PutDirectPlan500JSONResponse{
			Message: fmt.Sprintf("failed to commit transaction: %v", err),
		}
		return
	}

	if xmax == 0 {
		outResp = vcrest.PutDirectPlan201Response{}
	} else {
		outResp = vcrest.PutDirectPlan200Response{}
	}
	return
}
