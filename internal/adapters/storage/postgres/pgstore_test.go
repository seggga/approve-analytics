// Integration test, depends on running postgres instance
// started by command: make compose/up

package postgres

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/seggga/approve-analytics/internal/domain/models"
)

const (
	DSN = "postgres://root:pass@127.0.0.1:5432/test-db"
)

var (
	store *Store

	taskInsert = uint64(12)
	timeStamp  = time.Now()

	msgInsert = models.Message{
		EventType:  models.Created,
		TaskID:     taskInsert,
		Approver:   "",
		RecievedAt: timeStamp.Add(-99 * time.Second),
	}
	msgUpdate = models.Message{
		EventType:  models.MessageSent,
		TaskID:     taskInsert,
		Approver:   "approver@mail.com",
		RecievedAt: timeStamp.Add(-99 * time.Second),
	}
	msgUpdateDelay = models.Message{
		EventType:  models.Approved,
		TaskID:     taskInsert,
		Approver:   "approver@mail.com",
		RecievedAt: timeStamp.Add(-79 * time.Second),
	}

	msgInsForDelay []models.Message = []models.Message{
		{
			EventType:  models.MessageSent,
			TaskID:     101,
			Approver:   "approver101@mail.com",
			RecievedAt: timeStamp.Add(-100 * time.Second),
		},
		{
			EventType:  models.MessageSent,
			TaskID:     102,
			Approver:   "approver102@mail.com",
			RecievedAt: timeStamp.Add(-100 * time.Second),
		},
		{
			EventType:  models.MessageSent,
			TaskID:     103,
			Approver:   "approver103@mail.com",
			RecievedAt: timeStamp.Add(-100 * time.Second),
		},
		{
			EventType:  models.MessageSent,
			TaskID:     104,
			Approver:   "approver104@mail.com",
			RecievedAt: timeStamp.Add(-100 * time.Second),
		},
	}

	magUpdForDelay []models.Message = []models.Message{
		{
			EventType:  models.Finished,
			TaskID:     101,
			Approver:   "approver101@mail.com",
			RecievedAt: timeStamp.Add(-80 * time.Second),
		},
		{
			EventType:  models.Declined,
			TaskID:     102,
			Approver:   "approver102@mail.com",
			RecievedAt: timeStamp.Add(-70 * time.Second),
		},
		{
			EventType:  models.Finished,
			TaskID:     103,
			Approver:   "approver103@mail.com",
			RecievedAt: timeStamp.Add(-60 * time.Second),
		},
		{
			EventType:  models.Declined,
			TaskID:     104,
			Approver:   "approver104@mail.com",
			RecievedAt: timeStamp.Add(-50 * time.Second),
		},
	}
)

func TestMain(m *testing.M) {
	store, _ = New(DSN)
	_ = store.Init(context.TODO())

	defer clearDB()

	os.Exit(m.Run())
}

// test insert and select methods
func TestInsertSelect(t *testing.T) {
	// store, _ = New(DSN)

	err := store.Insert(context.TODO(), &msgInsert)
	if err != nil {
		t.Fatalf("unexpected error on insert: %v", err)
	}

	gotMsg, err := store.Select(context.TODO(), msgInsert.TaskID)
	if err != nil {
		t.Fatalf("unexpected error on select inserted msg %v: %v", msgInsert, err)
	}

	if gotMsg.Approver != msgInsert.Approver || gotMsg.EventType != msgInsert.EventType {
		t.Fatalf("wrong structure read: expected %v, got %v", msgInsert, gotMsg)
	}
}

// test approval sequence
func TestUpdate(t *testing.T) {
	// store, _ = New(DSN)

	err := store.Update(context.TODO(), &msgUpdate)
	if err != nil {
		t.Fatalf("unexpected error on update: %v", err)
	}

	gotMsg, err := store.Select(context.TODO(), msgUpdate.TaskID)
	if err != nil {
		t.Fatalf("unexpected error on select inserted msg %v: %v", msgUpdate, err)
	}

	if gotMsg.Approver != msgUpdate.Approver || gotMsg.EventType != msgUpdate.EventType {
		t.Fatalf("wrong structure read: expected %v, got %v", msgUpdate, gotMsg)
	}

}

func TestUpdateDelay(t *testing.T) {
	// store, _ = New(DSN)

	err := store.UpdateDelay(context.TODO(), &msgUpdateDelay)
	if err != nil {
		t.Fatalf("unexpected error on update: %v", err)
	}

	gotMsg, err := store.Select(context.TODO(), msgUpdateDelay.TaskID)
	if err != nil {
		t.Fatalf("unexpected error on select inserted msg %v: %v", msgUpdate, err)
	}

	if gotMsg.Approver != msgUpdateDelay.Approver || gotMsg.EventType != msgUpdateDelay.EventType {
		t.Fatalf("wrong structure read: expected %v, got %v", msgUpdateDelay, gotMsg)
	}

}
func TestGetAggregates(t *testing.T) {
	// store, _ = New(DSN)

	totalsExpected := models.Totals{
		Finished: 2,
		Declined: 2,
	}
	delaysExpected := []models.Delay{
		{
			ID:  101,
			Lag: time.Second * 20,
		},
		{
			ID:  102,
			Lag: time.Second * 30,
		},
		{
			ID:  103,
			Lag: time.Second * 40,
		},
		{
			ID:  104,
			Lag: time.Second * 50,
		},
	}

	ctx := context.TODO()
	for _, v := range msgInsForDelay {
		if err := store.Insert(ctx, &v); err != nil {
			t.Fatalf("error inserting messages, %v", err)
		}
	}

	for _, v := range magUpdForDelay {
		if err := store.UpdateDelay(ctx, &v); err != nil {
			t.Fatalf("error updating messages with delay , %v", err)
		}
	}

	totals, delays, err := store.GetAggregates(ctx)
	if err != nil {
		t.Fatalf("error getting statistics, %v", err)
	}

	if totalsExpected != *totals {
		t.Fatalf("wrong totals value: expected %v, got %v", totalsExpected, *totals)
	}

	// compare slices
	if delays == nil {
		t.Fatalf("got nil delays")
	}
	if len(delays) != len(delaysExpected) {
		t.Fatalf("got wrong delays len: expected %d, got %d", len(delaysExpected), len(delays))
	}

	if !reflect.DeepEqual(delays, delaysExpected) {
		t.Fatalf("got wrong delays: expected %v, got %v", delays, delaysExpected)
	}
}

func clearDB() {
	ctx := context.TODO()
	query := `
	DROP SCHEMA IF EXISTS analytics CASCADE;
	DROP TYPE IF EXISTS event_t CASCADE;
	`
	_, _ = store.Pool.Exec(ctx, query)
}
