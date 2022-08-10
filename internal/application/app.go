package application

import (
	"context"

	"github.com/seggga/approve-analytics/internal/adapters/auth"
	kfk "github.com/seggga/approve-analytics/internal/adapters/msglistener/kafkaconsumer"
	"github.com/seggga/approve-analytics/internal/adapters/rest"
	"github.com/seggga/approve-analytics/internal/adapters/storage/postgres"
	"github.com/seggga/approve-analytics/internal/domain/analytic"
	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"
)

var (
	restService *rest.Server
	msgListener *kfk.Client
	authClient  *auth.Client

	logger *zap.Logger
)

// Start ...
func Start(ctx context.Context) {

	var err error
	cfg := getConfig()
	logger := initLogger(cfg.Logger.Level)

	pgConn, err := postgres.New(cfg.Postgres.DSN)
	if err != nil {
		logger.Sugar().Fatalf("cannot connect to postgre: %v", err)
	}

	err = pgConn.Init(context.TODO())
	if err != nil {
		logger.Sugar().Fatalf("cannot init postgre schema: %v", err)
	}

	authClient, err = auth.NewClient(cfg.IFaces.AUTHAddress)
	if err != nil {
		logger.Sugar().Fatalf("cannot create gRPC client: %v", err)
	}
	analyticService := analytic.New(pgConn)
	restService = rest.New(logger, authClient, analyticService, cfg.IFaces.RESTPort)
	// msgListener = goodrpc.New(analytic.New(pgConn), logger, cfg.IFaces.MSGPort)
	msgListener, err = kfk.New(cfg.Kafka.Server, cfg.Kafka.Topic, cfg.Kafka.GroupID, logger, analyticService)
	if err != nil {
		logger.Sugar().Fatalf("cannot create kafka client: %v", err)
	}

	var g errgroup.Group
	g.Go(func() error {
		return restService.Start()
	})
	g.Go(func() error {
		return msgListener.Start(ctx)
	})

	logger.Info("app is started")
	err = g.Wait()
	if err != nil {
		logger.Sugar().Fatalf("http server start failed: %v", err)
	}

}

// Stop ...
func Stop() {
	defer logger.Sync()

	// stop REST
	err := restService.Stop(context.Background())
	if err != nil {
		logger.Sugar().Errorf("error stopping REST service", err)
	}
	// stop gRPC authentication client
	err = authClient.Conn.Close()
	if err != nil {
		logger.Sugar().Errorf("error stopping auth client service", err)
	}
	// stop kafka consumer
	err = msgListener.Stop()
	if err != nil {
		logger.Sugar().Errorf("error stopping kafka listener service", err)
	}

	logger.Sugar().Info("application has been stopped")

}
