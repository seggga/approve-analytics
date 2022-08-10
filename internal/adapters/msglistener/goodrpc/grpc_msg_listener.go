package goodrpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/seggga/approve-analytics/internal/domain/models"
	"github.com/seggga/approve-analytics/internal/ports"
	pb "github.com/seggga/approve-analytics/pkg/proto/analytics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	_ ports.MsgListener2 = &Server{}
)

// Server is a gRPC server based on pb package
type Server struct {
	an     ports.Analyter
	logger *zap.Logger

	srv      *grpc.Server
	listener net.Listener
	pb.UnimplementedAnalyticAPIServer
}

// New ..
func New(an ports.Analyter, logger *zap.Logger, port string) *Server {
	var err error
	s := &Server{
		an:     an,
		logger: logger,
		srv:    grpc.NewServer(),
	}
	s.listener, err = net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Sugar().Fatalf("error creating listener on port %s: %v", port, err)
	}
	pb.RegisterAnalyticAPIServer(s.srv, s)
	return s
}

// Start ...
func (s *Server) Start() error {
	s.logger.Debug("starting gRPC server ...")

	if err := s.srv.Serve(s.listener); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("cannot start gRPC server: %v", err)
	}
	return nil
}

// Stop ...
func (s *Server) Stop() {
	s.logger.Debug("stopping gRPC server ...")

	s.srv.GracefulStop()
	// s.logger.Info("gRPC server stopped")
}

// WriteMessage ...
func (s *Server) WriteMessage(ctx context.Context, req *pb.WriteMessageRequest) (*empty.Empty, error) {

	s.logger.Debug("incoming message...")

	msg := &models.Message{
		EventType:  req.EventType,
		TaskID:     req.TaskID,
		Approver:   req.Approver,
		RecievedAt: req.TimeStamp.AsTime(),
	}

	s.logger.Sugar().Debugf("message %v", msg)

	err := s.an.WriteEvent(ctx, msg)
	if err != nil {

		s.logger.Sugar().Debugf("error writing message: %v", err)
		return new(empty.Empty), fmt.Errorf("error writing message: %v", req)
	}

	s.logger.Debug("message has been written")

	return new(empty.Empty), nil
}
