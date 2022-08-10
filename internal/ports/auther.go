package ports

import (
	"context"

	"github.com/seggga/approve-analytics/internal/domain/models"
)

// Auther ...
type Auther interface {
	Authenticate(ctx context.Context, token *models.TokenPair) (*models.TokenPair, error)
	// Validate(ctx context.Context, tokens *models.TokenPair) error
}
