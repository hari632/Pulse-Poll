package models

import "time"

type Vote struct {
	ID        string    `json:"id" bson:"_id"`
	PollID    string    `json:"pollId" bson:"pollId"`
	OptionID  string    `json:"optionId" bson:"optionId"`
	VoterID   string    `json:"voterId,omitempty" bson:"voterId,omitempty"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}