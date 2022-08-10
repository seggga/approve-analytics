package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/seggga/approve-analytics/internal/domain/models"
)

const (
	DDL = `

	DROP SCHEMA IF EXISTS analytics CASCADE;
	DROP TYPE IF EXISTS event_t CASCADE;

	CREATE SCHEMA IF NOT EXISTS analytics;
	CREATE TYPE event_t AS enum
	(
		'CREATED',
		'MESSAGE_SENT',
		'APPROVED',
		'DECLINED',
		'FINISHED',
		'DELETED'
	);	
	CREATE TABLE IF NOT EXISTS analytics.events
	(
		id serial4 NOT NULL,
		task_id INT4 NOT NULL,
		event_type event_t NOT NULL,
		approver_email varchar(256) NOT NULL,
		recieved_at timestamp with time zone NOT NULL,
		total_delay interval SECOND DEFAULT NULL,
	
		CONSTRAINT events_pkey PRIMARY KEY (id)
	);
	CREATE TABLE IF NOT EXISTS analytics.totals
	(
		id INT DEFAULT 0,
		finished INT4,
		declined INT4
	);
	`
)

// Store ...
type Store struct {
	Pool *pgxpool.Pool
}

// Init creates schema, type and tables
func (s *Store) Init(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, DDL)
	return err
}

// New ...
func New(dsn string) (s *Store, err error) {
	ctx := context.Background()

	pool, err := pgxpool.Connect(context.TODO(), dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}
	s = &Store{Pool: pool}

	return s, nil
}

// Insert adds event about task that has not been stored yet
func (s *Store) Insert(ctx context.Context, msg *models.Message) error {
	query := "INSERT INTO analytics.events (task_id, event_type, approver_email, recieved_at, total_delay) values ($1, $2, $3, $4, interval '0 second') RETURNING id"
	row := s.Pool.QueryRow(ctx, query,
		msg.TaskID,
		msg.EventType,
		msg.Approver,
		msg.RecievedAt,
	)

	var id uint64
	if err := row.Scan(&id); err != nil {
		return fmt.Errorf("error inserting new event in db: %v", err)
	}

	return nil
}

// Select extracts an event with specified ID
func (s *Store) Select(ctx context.Context, taskID uint64) (*models.Message, error) {
	evt := &models.Event{}
	query := "SELECT event_type, task_id, approver_email, recieved_at FROM analytics.events WHERE task_id=$1"
	err := s.Pool.QueryRow(ctx, query, taskID).Scan(&evt.EventType, &evt.TaskID, &evt.Approver, &evt.RecievedAt)

	// ErrNoRows is a handled situation meaning a massage with a new task is received
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return &models.Message{
		EventType:  evt.EventType,
		TaskID:     evt.TaskID,
		Approver:   evt.Approver,
		RecievedAt: evt.RecievedAt,
	}, nil
}

// Update changes event about particular task in database with msg values
func (s *Store) Update(ctx context.Context, msg *models.Message) error {
	query := "UPDATE analytics.events SET event_type=$1, approver_email=$2, recieved_at=$3 WHERE task_id=$4"
	_, err := s.Pool.Exec(ctx, query, msg.EventType, msg.Approver, msg.RecievedAt, msg.TaskID)
	return err
}

// UpdateDelay sets data about particular task in database with msg values and calculated delay
func (s *Store) UpdateDelay(ctx context.Context, msg *models.Message) error {
	var (
		timeStamp time.Time
		delay     time.Duration
	)
	query := "SELECT recieved_at, total_delay FROM analytics.events WHERE task_id=$1"
	err := s.Pool.QueryRow(ctx, query, msg.TaskID).Scan(&timeStamp, &delay)
	if err != nil {
		return err
	}

	duration := msg.RecievedAt.Sub(timeStamp) + delay
	query = "UPDATE analytics.events SET event_type=$1, approver_email=$2, recieved_at=$3, total_delay=$4 WHERE task_id=$5"
	_, err = s.Pool.Exec(ctx, query, msg.EventType, msg.Approver, msg.RecievedAt, duration, msg.TaskID)
	return err
}

// GetAggregates extracts statistics about finished and declined tasks and its delay
func (s *Store) GetAggregates(ctx context.Context) (*models.Totals, []models.Delay, error) {

	// refresh totals in DB
	err := s.calculateAggregates(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error calculating aggregates, %v", err)
	}

	var (
		totals models.Totals
		delay  models.Delay
	)
	// get totals
	query := `SELECT finished, declined FROM analytics.totals as t WHERE t.id=0;`
	if err := s.Pool.QueryRow(ctx, query).Scan(&totals.Finished, &totals.Declined); err != nil {
		return nil, nil, fmt.Errorf("error reading totals (finished and declined tasks): %v", err)
	}

	// get delays
	query = `SELECT task_id, total_delay t_delay FROM analytics.events e WHERE e.total_delay IS NOT NULL AND e.event_type in ('DECLINED', 'FINISHED', 'DELETED');`
	rows, err := s.Pool.Query(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("error selecting task delays: %v", err)
	}
	defer rows.Close()

	delays := make([]models.Delay, 0, totals.Finished+totals.Declined)
	for rows.Next() {
		err = rows.Scan(&delay.ID, &delay.Lag)
		if err != nil {
			return nil, nil, fmt.Errorf("error reading task delays: %v", err)
		}
		delays = append(delays, delay)
	}

	// TODO: change to separate table and simplify this method to jush read values
	return &totals, delays, nil
}

// calculateAggregates updates data in analytics.totals counting FINISHED and DECLINED
// tasks. Also counts number of not nil delays to pass to caller function.
// Queries are executed in a transaction
func (s *Store) calculateAggregates(ctx context.Context) error {

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error on transaction begin, %v", err)
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	defer tx.Rollback(context.Background())

	query := `INSERT INTO analytics.totals (id, finished, declined) VALUES (0,0,0);`
	_, err = tx.Exec(ctx, query)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("error inserting 0-values in totals, %v", err)
	}

	query = `UPDATE analytics.totals t SET (finished, declined) = (
		(select  Sum(case when event_type = 'FINISHED' then 1 else 0 end) from analytics.events),
		(select  Sum(case when event_type in ('DECLINED', 'DELETED') then 1 else 0 end) from analytics.events)
	)
	WHERE t.id=0;`
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("error updating totals, %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error on transaction commit, %v", err)
	}

	return nil
}

// Drop clears database (for testing purpose onle)
func (s *Store) Drop(ctx context.Context) error {
	query := `
DROP TYPE IF EXISTS event_t CASCADE;
DROP SCHEMA IF EXISTS analytics CASCADE;
`
	_, err := s.Pool.Exec(ctx, query)

	return err
}
