// Integration test, depends on running postgres instance
// started by command: make compose/up

package goodrpc

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/seggga/approve-analytics/internal/adapters/storage/postgres"
	"github.com/seggga/approve-analytics/internal/domain/analytic"
	"github.com/seggga/approve-analytics/internal/domain/models"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/seggga/approve-analytics/pkg/proto/analytics"
)

const (
	DSN = "postgres://root:pass@127.0.0.1:5432/test-db"
)

var (
	s     *Server
	store *postgres.Store

	timeStamp    = time.Now()
	taskFinished = uint64(120)
	taskDeclined = uint64(121)

	messages []models.Message = []models.Message{
		// approved task
		{
			EventType:  models.Created,
			TaskID:     taskFinished,
			Approver:   "",
			RecievedAt: timeStamp.Add(-100 * time.Second),
		},
		{
			EventType:  models.MessageSent,
			TaskID:     taskFinished,
			Approver:   "approver110@mail.com",
			RecievedAt: timeStamp.Add(-99 * time.Second),
		},
		{
			EventType:  models.Approved,
			TaskID:     taskFinished,
			Approver:   "approver110@mail.com",
			RecievedAt: timeStamp.Add(-89 * time.Second), // 10 seconds
		},
		{
			EventType:  models.MessageSent,
			TaskID:     taskFinished,
			Approver:   "approver111@mail.com",
			RecievedAt: timeStamp.Add(-89 * time.Second),
		},
		{
			EventType:  models.Approved,
			TaskID:     taskFinished,
			Approver:   "approver111@mail.com",
			RecievedAt: timeStamp.Add(-19 * time.Second), // 70 seconds
		},
		{
			EventType:  models.Finished,
			TaskID:     taskFinished,
			Approver:   "",
			RecievedAt: timeStamp.Add(-18 * time.Second),
		},

		// declined task
		{
			EventType:  models.Created,
			TaskID:     taskDeclined,
			Approver:   "",
			RecievedAt: timeStamp.Add(-100 * time.Second),
		},
		{
			EventType:  models.MessageSent,
			TaskID:     taskDeclined,
			Approver:   "approver120@mail.com",
			RecievedAt: timeStamp.Add(-99 * time.Second),
		},
		{
			EventType:  models.Approved,
			TaskID:     taskDeclined,
			Approver:   "approver120@mail.com",
			RecievedAt: timeStamp.Add(-79 * time.Second), // 20 seconds
		},
		{
			EventType:  models.MessageSent,
			TaskID:     taskDeclined,
			Approver:   "approver121@mail.com",
			RecievedAt: timeStamp.Add(-79 * time.Second),
		},
		{
			EventType:  models.Declined,
			TaskID:     taskDeclined,
			Approver:   "approver121@mail.com",
			RecievedAt: timeStamp.Add(-19 * time.Second), // 60 seconds
		},
	}

	totalsExpected = models.Totals{
		Finished: 1,
		Declined: 1,
	}
	delaysExpected = []models.Delay{
		{
			ID:  120,
			Lag: time.Second * 20,
		},
		{
			ID:  121,
			Lag: time.Second * 30,
		},
	}
)

func TestMain(m *testing.M) {
	store, _ = postgres.New(DSN)
	err := store.Init(context.TODO())
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	logger, _ := zap.NewDevelopment()
	s = New(analytic.New(store), logger, "4000")
	go s.Start()
	os.Exit(m.Run())
}

type Client struct {
	conn *grpc.ClientConn
	pb.AnalyticAPIClient
}

func TestWriteEvent(t *testing.T) {

	// create grpc client
	path := "127.0.0.1:4000"
	ctx := context.TODO()
	conn, err := grpc.DialContext(ctx, path, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Fatalf("error creating client connection %s: %v", path, err)
	}
	client := pb.NewAnalyticAPIClient(conn)

	cl := Client{
		conn:              conn,
		AnalyticAPIClient: client,
	}

	defer cl.conn.Close()

	// send messages
	for _, v := range messages {
		msgReq := &pb.WriteMessageRequest{
			EventType: v.EventType,
			TaskID:    v.TaskID,
			Approver:  v.Approver,
			TimeStamp: timestamppb.New(v.RecievedAt),
		}
		if _, err := cl.WriteMessage(ctx, msgReq); err != nil {
			t.Fatalf("error on message %v: %v", v, err)
		}
	}

	// read result from DB
	totalsExpect := models.Totals{Finished: 1, Declined: 1}
	delaysExpect := []models.Delay{
		{
			ID:  taskFinished,
			Lag: time.Second * 80,
		},
		{
			ID:  taskDeclined,
			Lag: time.Second * 80,
		},
	}

	totals, delays, err := store.GetAggregates(ctx)
	if err != nil {
		t.Fatalf("unexpected error on getting aggregates. %v", err)
	}

	if *totals != totalsExpect {
		t.Fatalf("wrong totals, expected %v, got %v", totalsExpect, totals)
	}

	if !reflect.DeepEqual(delays, delaysExpect) {
		t.Fatalf("wrong delays, expected %v, got %v", delaysExpect, delays)
	}

	s.Stop()

}
