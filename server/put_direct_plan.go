package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
)

// PutDirectPlan adds or updates a direct plan with the given UUID
func (s *Server) PutDirectPlan(ctx context.Context, request vcrest.PutDirectPlanRequestObject) (outResp vcrest.PutDirectPlanResponseObject, _ error) {
	// Validate request.
	requestUuid, err := internal.AsUUID(request.Uuid)
	if err != nil {
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: "invalid UUID format",
		}
		return
	}
	if requestUuid == uuid.Nil {
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: "UUID cannot be zero",
		}
		return
	}
	if request.Body == nil {
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: "request body is required",
		}
		return
	}
	if err := errors.Join(
		internal.FieldRequired(request.Body.SourceUuid),
		internal.FieldNotNull(request.Body.SourceUuid),
		internal.FieldValidUUID(request.Body.SourceUuid),
	); err != nil {
		outResp = vcrest.PutDirectPlan400JSONResponse{
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
		outResp = vcrest.PutDirectPlan400JSONResponse{
			Message: fmt.Sprintf("WorkUuid: %v", err),
		}
		return
	}

	workUuid := internal.FieldMustUUID(request.Body.WorkUuid)

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

	result, err := internal.UpsertEntity(ctx, txn, "plans", requestUuid, internal.PlanKindDirect, bodyRaw)
	if errors.Is(err, internal.ErrUpsertType) {
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

	if err := internal.UpdatePlanInputs(ctx, txn, requestUuid, sourceUuid); err != nil {
		outResp = vcrest.PutDirectPlan500JSONResponse{
			Message: err.Error(),
		}
		return
	}

	if err := internal.UpdatePlanOutputs(ctx, txn, requestUuid, workUuid); err != nil {
		outResp = vcrest.PutDirectPlan500JSONResponse{
			Message: err.Error(),
		}
		return
	}

	if err := txn.Commit(ctx); err != nil {
		outResp = vcrest.PutDirectPlan500JSONResponse{
			Message: fmt.Sprintf("failed to commit transaction: %v", err),
		}
		return
	}

	if result == internal.UpsertCreated {
		outResp = vcrest.PutDirectPlan201Response{}
	} else {
		outResp = vcrest.PutDirectPlan200Response{}
	}
	return
}
