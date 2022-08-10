package models

import "time"

// Message represents incoming message from Task and Mail services
type Message struct {
	EventType  string    `json:"eventtype"`
	TaskID     uint64    `json:"taskid"`
	Approver   string    `json:"approver"`
	RecievedAt time.Time `json:"recievedat"`
}
