package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UpsertResult indicates whether an upsert created a new row or updated an existing one.
type UpsertResult bool

const (
	UpsertCreated UpsertResult = true
	UpsertUpdated UpsertResult = false
)

// ErrUpsertType is returned when an upsert fails because the entity exists with a different type/kind.
var ErrUpsertType = errors.New("entity exists with different type")

// Querier is an interface that can execute QueryRow, implemented by both pgxpool.Pool and pgx.Tx.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Execer is an interface that can execute Exec, implemented by both pgxpool.Pool and pgx.Tx.
type Execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// UpsertEntity performs an INSERT ON CONFLICT DO UPDATE for an entity table (works, sources, plans).
// Returns UpsertCreated if a new row was created, UpsertUpdated if an existing row was updated,
// or ErrUpsertType if the row exists with a different kind.
func UpsertEntity(ctx context.Context, q Querier, table string, id uuid.UUID, kind string, body json.RawMessage) (UpsertResult, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (uuid, kind, body)
		VALUES ($1, $2, $3)
		ON CONFLICT (uuid) DO UPDATE
		SET body = EXCLUDED.body
		WHERE %s.kind = $2
		RETURNING xmax`, table, table)

	var xmax uint32
	err := q.QueryRow(ctx, query, id, kind, body).Scan(&xmax)
	if errors.Is(err, pgx.ErrNoRows) {
		return UpsertUpdated, ErrUpsertType
	} else if err != nil {
		return UpsertUpdated, fmt.Errorf("failed to upsert %s: %w", table, err)
	}

	if xmax == 0 {
		return UpsertCreated, nil
	}
	return UpsertUpdated, nil
}

// UpdatePlanInputs replaces the plan_inputs entries for a plan with a single source UUID.
func UpdatePlanInputs(ctx context.Context, tx pgx.Tx, planUUID, sourceUUID uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM plan_inputs WHERE plan_uuid = $1`, planUUID)
	if err != nil {
		return fmt.Errorf("failed to delete old plan_inputs: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO plan_inputs (plan_uuid, source_uuid) VALUES ($1, $2)`, planUUID, sourceUUID)
	if err != nil {
		return fmt.Errorf("failed to insert plan_inputs: %w", err)
	}
	return nil
}

// UpdatePlanOutputs replaces the plan_outputs entries for a plan with a single work UUID.
func UpdatePlanOutputs(ctx context.Context, tx pgx.Tx, planUUID, workUUID uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM plan_outputs WHERE plan_uuid = $1`, planUUID)
	if err != nil {
		return fmt.Errorf("failed to delete old plan_outputs: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO plan_outputs (plan_uuid, work_uuid) VALUES ($1, $2)`, planUUID, workUUID)
	if err != nil {
		return fmt.Errorf("failed to insert plan_outputs: %w", err)
	}
	return nil
}

// NewDBPool creates a new pgxpool.Pool from the given DatabaseConfig.
func NewDBPool(ctx context.Context, cfg *DatabaseConfig) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
