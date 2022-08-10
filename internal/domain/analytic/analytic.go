package analytic

import (
	"context"
	"fmt"

	"github.com/seggga/approve-analytics/internal/domain/models"
	"github.com/seggga/approve-analytics/internal/ports"
)

var (
	_ ports.Analyter = &Service{}
)

// Service implements main analytics logic
type Service struct {
	db ports.EventStorage
}

// New creates a new auth service
func New(db ports.EventStorage) *Service {
	return &Service{
		db: db,
	}
}

// WriteEvent receives a Message from message service.
// Like a state-machine WriteEvent has finite number of transitions
// for a particuar task:
//
// no task_id 	-> CREATED
// CREATED 		-> MESSAGE_SENT
// MESSAGE_SENT -> ( APPROVED || DECLINED ) && ( approver == approver )
// APPROVED 	-> FINISHED || MESSAGE_SENT
//
// CREATED || MESSAGE_SENT || APPROVED -> DELETED
func (s *Service) WriteEvent(ctx context.Context, msg *models.Message) error {
	evt, err := s.db.Select(ctx, msg.TaskID)
	if err != nil {
		return fmt.Errorf("error selecting event by taskID, %v, %v", msg, err)
	}
	// no task_id 	-> CREATED
	if evt == nil && msg.EventType == models.Created {
		err = s.db.Insert(ctx, msg)
		if err != nil {
			return fmt.Errorf("error writing event data into storage: %v, %v", msg, err)
		}

		return nil
	}

	// CREATED 	-> MESSAGE_SENT
	if evt != nil && evt.EventType == models.Created && msg.EventType == models.MessageSent {
		err = s.db.Update(ctx, msg)
		if err != nil {
			return fmt.Errorf("error updating event data in storage: %v, %v, %v", evt, msg, err)
		}
		return nil
	}

	// MESSAGE_SENT -> ( APPROVED || DECLINED ) && ( approver == approver )
	if evt != nil && evt.EventType == models.MessageSent && (msg.EventType == models.Approved || msg.EventType == models.Declined) {
		if evt.Approver != msg.Approver {
			return fmt.Errorf("approvers are not equal: %v %v", evt, msg)
		}
		err = s.db.UpdateDelay(ctx, msg)
		if err != nil {
			return fmt.Errorf("error updating event with delay in storage: %v, %v, %v", evt, msg, err)
		}
		return nil
	}

	// APPROVED 	-> FINISHED || MESSAGE_SENT
	if evt != nil && evt.EventType == models.Approved && (msg.EventType == models.Finished || msg.EventType == models.MessageSent) {
		err = s.db.Update(ctx, msg)
		if err != nil {
			return fmt.Errorf("error updating event data in storage: %v, %v, %v", evt, msg, err)
		}
		return nil
	}

	// CREATED || APPROVED -> DELETED
	// no additional delay
	// no matter, who is approver - if Tasks service sent DELETE message,
	// than means the sender is the task owner
	if evt != nil && (evt.EventType == models.Created || evt.EventType == models.Approved) && msg.EventType == models.Deleted {
		err = s.db.Update(ctx, msg)
		if err != nil {
			return fmt.Errorf("error updating event data in storage: %v, %v, %v", evt, msg, err)
		}
		return nil
	}

	// MESSAGE_SENT -> DELETED
	// calculates additional delay
	// no matter, who is approver - if Tasks service sent DELETE message,
	// than means the sender is the task owner
	if evt != nil && evt.EventType == models.MessageSent && msg.EventType == models.Deleted {
		err = s.db.UpdateDelay(ctx, msg)
		if err != nil {
			return fmt.Errorf("error updating event with delay in storage: %v, %v, %v", evt, msg, err)
		}
		return nil
	}

	return fmt.Errorf("due to previous found event %v, message %v has not been classified as valid", evt, msg)
}

// GetAggregates extracts requested data
func (s *Service) GetAggregates(ctx context.Context) (*models.Totals, []models.Delay, error) {

	totals, delays, err := s.db.GetAggregates(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting aggregates from DB, %v", err)
	}

	return totals, delays, nil
}
