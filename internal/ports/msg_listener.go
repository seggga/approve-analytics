package ports

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/seggga/approve-analytics/internal/domain/models"
	pb "github.com/seggga/approve-analytics/pkg/proto/analytics"
)

// MsgListener2 receives messages from other services via gRPC
type MsgListener2 interface {
	WriteMessage(ctx context.Context, r *pb.WriteMessageRequest) (*empty.Empty, error)
}

// MsgListener a universal interface for message listener
type MsgListener interface {
	ProcessMessage(ctx context.Context, msg *models.Message) error
}
