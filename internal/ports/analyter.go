package ports

import (
	"context"

	"github.com/seggga/approve-analytics/internal/domain/models"
)

// Analyter ...
type Analyter interface {
	WriteEvent(ctx context.Context, msg *models.Message) error
	GetAggregates(ctx context.Context) (*models.Totals, []models.Delay, error)

	// Authenticate(ctx context.Context, tokens *models.TokenPair) (*models.TokenPair, error)
}
