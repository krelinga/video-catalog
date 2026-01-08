package main

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/krelinga/video-catalog/internal"
	"github.com/krelinga/video-catalog/vcrest"
	"github.com/oapi-codegen/nullable"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	pageTokenMagic  = uint32(0x504c414e) // "PLAN" in ASCII
	defaultPageSize = 50
	minPageSize     = 1
	maxPageSize     = 500
)

type pageToken struct {
	Magic    uint32
	LastUUID uuid.UUID
}

func encodePageToken(lastUUID uuid.UUID) (string, error) {
	token := pageToken{
		Magic:    pageTokenMagic,
		LastUUID: lastUUID,
	}
	buf := make([]byte, 4+16) // 4 bytes for magic + 16 bytes for UUID
	binary.BigEndian.PutUint32(buf[0:4], token.Magic)
	copy(buf[4:], token.LastUUID[:])
	return base64.URLEncoding.EncodeToString(buf), nil
}

func decodePageToken(tokenStr string) (uuid.UUID, error) {
	buf, err := base64.URLEncoding.DecodeString(tokenStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid page token encoding: %w", err)
	}
	if len(buf) != 20 {
		return uuid.Nil, fmt.Errorf("invalid page token length: expected 20, got %d", len(buf))
	}
	magic := binary.BigEndian.Uint32(buf[0:4])
	if magic != pageTokenMagic {
		return uuid.Nil, fmt.Errorf("invalid page token magic: expected 0x%08x, got 0x%08x", pageTokenMagic, magic)
	}
	var lastUUID uuid.UUID
	copy(lastUUID[:], buf[4:])
	return lastUUID, nil
}

// ListPlans lists plans with optional filtering.
func (s *Server) ListPlans(ctx context.Context, request vcrest.ListPlansRequestObject) (outResp vcrest.ListPlansResponseObject, _ error) {
	// Determine page size with reasonable bounds
	pageSize := defaultPageSize
	if request.Params.PageSize != nil {
		ps := *request.Params.PageSize
		if ps < minPageSize {
			pageSize = minPageSize
		} else if ps > maxPageSize {
			pageSize = maxPageSize
		} else {
			pageSize = int(ps)
		}
	}

	// Decode page token if provided
	var lastUUID uuid.UUID
	if request.Params.PageToken != nil && *request.Params.PageToken != "" {
		var err error
		lastUUID, err = decodePageToken(*request.Params.PageToken)
		if err != nil {
			outResp = vcrest.ListPlans400JSONResponse{
				Message: fmt.Sprintf("invalid page token: %v", err),
			}
			return
		}
	}

	// Parse optional filter UUIDs
	var sourceUUID uuid.UUID
	var workUUID uuid.UUID
	if request.Params.SourceUuid != nil {
		var err error
		sourceUUID, err = uuid.Parse(request.Params.SourceUuid.String())
		if err != nil {
			outResp = vcrest.ListPlans400JSONResponse{
				Message: "invalid sourceUuid format",
			}
			return
		}
	}
	if request.Params.WorkUuid != nil {
		var err error
		workUUID, err = uuid.Parse(request.Params.WorkUuid.String())
		if err != nil {
			outResp = vcrest.ListPlans400JSONResponse{
				Message: "invalid workUuid format",
			}
			return
		}
	}

	txn, err := s.Pool.Begin(ctx)
	if err != nil {
		outResp = vcrest.ListPlans500JSONResponse{
			Message: fmt.Sprintf("failed to begin transaction: %v", err),
		}
		return
	}
	defer txn.Rollback(ctx)

	// Build query with optional joins and filters
	query := `
		SELECT DISTINCT p.uuid, p.kind, p.body
		FROM plans p`

	args := []any{}
	argIdx := 1
	whereConditions := []string{}

	// Add join for source UUID filter
	if sourceUUID != uuid.Nil {
		query += `
		INNER JOIN plan_inputs pi ON p.uuid = pi.plan_uuid`
		whereConditions = append(whereConditions, fmt.Sprintf("pi.source_uuid = $%d", argIdx))
		args = append(args, sourceUUID)
		argIdx++
	}

	// Add join for work UUID filter
	if workUUID != uuid.Nil {
		query += `
		INNER JOIN plan_outputs po ON p.uuid = po.plan_uuid`
		whereConditions = append(whereConditions, fmt.Sprintf("po.work_uuid = $%d", argIdx))
		args = append(args, workUUID)
		argIdx++
	}

	// Add pagination filter
	if lastUUID != uuid.Nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.uuid > $%d", argIdx))
		args = append(args, lastUUID)
		argIdx++
	}

	// Add WHERE clause if we have conditions
	if len(whereConditions) > 0 {
		query += "\n\t\tWHERE "
		for i, cond := range whereConditions {
			if i > 0 {
				query += " AND "
			}
			query += cond
		}
	}

	// Add ordering and limit
	query += fmt.Sprintf(`
		ORDER BY p.uuid
		LIMIT $%d`, argIdx)
	args = append(args, pageSize+1) // Fetch one extra to determine if there's a next page

	plans := []vcrest.Plan{}
	var nextPageLastUUID uuid.UUID
	hasMore := false

	type planRow struct {
		uuid    uuid.UUID
		kind    internal.PlanKind
		bodyRaw json.RawMessage
	}

	var row planRow
	rows, err := txn.Query(ctx, query, args...)
	if err != nil {
		outResp = vcrest.ListPlans500JSONResponse{
			Message: fmt.Sprintf("failed to query plans: %v", err),
		}
		return
	}

	_, err = pgx.ForEachRow(rows, []any{&row.uuid, &row.kind, &row.bodyRaw}, func() error {
		if len(plans) >= pageSize {
			hasMore = true
			return nil
		}

		if !row.kind.IsValid() {
			return fmt.Errorf("invalid plan kind in database: %s", row.kind)
		}

		plan := vcrest.Plan{
			Uuid: openapi_types.UUID(row.uuid),
		}

		switch row.kind {
		case internal.PlanKindDirect:
			var directBody internal.DirectPlan
			if err := json.Unmarshal(row.bodyRaw, &directBody); err != nil {
				return fmt.Errorf("failed to unmarshal direct plan body: %w", err)
			}
			plan.Direct = &vcrest.DirectPlan{
				SourceUuid: nullable.NewNullableWithValue(openapi_types.UUID(directBody.SourceUUID)),
				WorkUuid:   nullable.NewNullableWithValue(openapi_types.UUID(directBody.WorkUUID)),
			}
		case internal.PlanKindChapterRange:
			var chapterRangeBody internal.ChapterRangePlan
			if err := json.Unmarshal(row.bodyRaw, &chapterRangeBody); err != nil {
				return fmt.Errorf("failed to unmarshal chapter range plan body: %w", err)
			}
			plan.ChapterRange = &vcrest.ChapterRangePlan{
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
			}
		default:
			return fmt.Errorf("unimplemented plan kind: %s", row.kind)
		}

		plans = append(plans, plan)
		nextPageLastUUID = row.uuid
		return nil
	})
	if err != nil {
		outResp = vcrest.ListPlans500JSONResponse{
			Message: fmt.Sprintf("failed to query and scan plans: %v", err),
		}
		return
	}

	if err := txn.Commit(ctx); err != nil {
		outResp = vcrest.ListPlans500JSONResponse{
			Message: fmt.Sprintf("failed to commit transaction: %v", err),
		}
		return
	}

	// Build response
	response := vcrest.ListPlans200JSONResponse{
		Plans: plans,
	}

	// Add next page token if there are more results
	if hasMore && nextPageLastUUID != uuid.Nil {
		token, err := encodePageToken(nextPageLastUUID)
		if err != nil {
			outResp = vcrest.ListPlans500JSONResponse{
				Message: fmt.Sprintf("failed to encode page token: %v", err),
			}
			return
		}
		response.NextPageToken = &token
	}

	outResp = response
	return
}
