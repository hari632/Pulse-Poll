package models

import "time"

type PollOption struct {
	ID   string `json:"id" bson:"id"`
	Text string `json:"text" bson:"text"`
}

type Poll struct {
	ID              string       `json:"id" bson:"_id"`
	Code            string       `json:"code,omitempty" bson:"code,omitempty"`
	CreatorID       string       `json:"creatorId" bson:"creatorId"`
	Question        string       `json:"question" bson:"question"`
	Options         []PollOption `json:"options" bson:"options"`
	Status          string       `json:"status" bson:"status"`
	AllowVoteChange bool         `json:"allowVoteChange" bson:"allowVoteChange"`
	Anonymous       bool         `json:"anonymous" bson:"anonymous"`
	ExpiresAt       *time.Time   `json:"expiresAt,omitempty" bson:"expiresAt,omitempty"`
	CreatedAt       time.Time    `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt" bson:"updatedAt"`
}

type PollResponseDTO struct {
	ID          string       `json:"id"`
	Code        string       `json:"code"`
	Question    string       `json:"question"`
	Options     []string     `json:"options"`
	RawOptions  []PollOption `json:"rawOptions,omitempty"`
	Votes       []int        `json:"votes"`
	TotalVotes  int          `json:"totalVotes"`
	Percentages []int        `json:"percentages"`
	Status      string       `json:"status"`
	CreatorID   string       `json:"creatorId,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
}