// Integration test, depends on running kafka, zookeeper and postgres instances
// started by command: make kafka/compose/up

package kafkaconsumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/seggga/approve-analytics/internal/adapters/storage/postgres"
	"github.com/seggga/approve-analytics/internal/domain/analytic"
	"github.com/seggga/approve-analytics/internal/domain/models"
	"go.uber.org/zap"

	"github.com/segmentio/kafka-go"
)

const (
	DSN = "postgres://root:pass@127.0.0.1:5432/test-db"

	broker  = "127.0.0.1:9093"
	topic   = "test-topic"
	groupID = "test-consumer-group"
)

var (
	c     *Client
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
	c, err = New(broker, topic, groupID, logger, analytic.New(store))
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	defer c.Reader.Close()

	os.Exit(m.Run())
}

func TestClient(t *testing.T) {
	type publisher struct {
		writer *kafka.Writer
	}

	pub := publisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}

	// convert simple messages to []byte to compose kafka messages
	kfkMessages := make([]kafka.Message, 0, len(messages))
	for _, m := range messages {
		value, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("cannot marshal msg %v, %v", m, err)
		}

		kfkMsg := kafka.Message{
			Value: value,
		}
		kfkMessages = append(kfkMessages, kfkMsg)
	}

	// send test messages to kafka.
	err := pub.writer.WriteMessages(context.Background(), kfkMessages...)
	if err != nil {
		t.Fatal(err)
	}
	pub.writer.Close()
	t.Log("messages sent to kafka")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	c.Start(ctx)
	<-ctx.Done()
	// read result from DB
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
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

	c.Stop()
}
