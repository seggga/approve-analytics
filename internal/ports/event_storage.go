package ports

import (
	"context"

	"github.com/seggga/approve-analytics/internal/domain/models"
)

// EventStorage ...
type EventStorage interface {
	Insert(ctx context.Context, msg *models.Message) error
	Select(ctx context.Context, ID uint64) (*models.Message, error)
	Update(ctx context.Context, msg *models.Message) error
	UpdateDelay(ctx context.Context, msg *models.Message) error

	GetAggregates(ctx context.Context) (*models.Totals, []models.Delay, error)
}
