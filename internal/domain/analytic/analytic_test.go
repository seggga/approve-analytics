package analytic

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/seggga/approve-analytics/internal/adapters/storage/postgres"
	"github.com/seggga/approve-analytics/internal/domain/models"
)

const (
	DSN = "postgres://root:pass@127.0.0.1:5432/test-db"
)

var (
	an Service

	timeStamp    = time.Now()
	taskFinished = uint64(110)
	taskDeclined = uint64(111)

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
)

func TestMain(m *testing.M) {
	store, _ := postgres.New(DSN)
	err := store.Init(context.TODO())
	if err != nil {
		os.Exit(2)
	}
	an.db = store

	os.Exit(m.Run())
}

func TestWriteEvent(t *testing.T) {
	ctx := context.TODO()
	for _, v := range messages {
		if err := an.WriteEvent(ctx, &v); err != nil {
			t.Fatalf("unexpected error on message %v: %v", v, err)
		}
	}
}

func TestGetAggregates(t *testing.T) {
	ctx := context.TODO()

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

	totals, delays, err := an.GetAggregates(ctx)
	if err != nil {
		t.Fatalf("unexpected error on getting aggregates. %v", err)
	}

	if *totals != totalsExpect {
		t.Fatalf("wrong totals, expected %v, got %v", totalsExpect, totals)
	}

	if !reflect.DeepEqual(delays, delaysExpect) {
		t.Fatalf("wrong delays, expected %v, got %v", delaysExpect, delays)
	}
}
