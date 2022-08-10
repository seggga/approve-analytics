package models

import "time"

// all possible event Statuses
const (
	Created     string = "CREATED"
	MessageSent string = "MESSAGE_SENT"
	Approved    string = "APPROVED"
	Declined    string = "DECLINED"
	Finished    string = "FINISHED"
	Deleted     string = "DELETED"
)

// Event represents a struct to store in database
type Event struct {
	ID         uint64
	EventType  string
	TaskID     uint64
	Approver   string
	RecievedAt time.Time
	Delay      time.Duration
}
