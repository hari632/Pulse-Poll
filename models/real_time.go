package models

type VoteEvent struct {
	PollID      string  `json:"pollId"`
	PollCode    string  `json:"pollCode,omitempty"`
	OptionID    string  `json:"optionId"`
	OptionIndex int     `json:"optionIndex"`
	Votes       []int64 `json:"votes"`
	OptionVotes int64   `json:"optionVotes"`
	TotalVotes  int64   `json:"totalVotes"`
	Percentages []int   `json:"percentages"`
}