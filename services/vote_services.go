package services

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"pulse-backend/models"
	"pulse-backend/repositories"
)

type VoteService struct {
	pollRepository repositories.PollRepository
	voteRepository repositories.VoteRepository
	voteCounter    *repositories.RedisVoteCounter
	eventPublisher *repositories.RedisEventPublisher
}

type VoteResult struct {
	OptionID   string  `json:"optionId"`
	OptionText string  `json:"optionText"`
	Votes      int     `json:"votes"`
	Percentage float64 `json:"percentage"`
}

func NewVoteService(
	pollRepository repositories.PollRepository,
	voteRepository repositories.VoteRepository,
	voteCounter *repositories.RedisVoteCounter,
	eventPublisher *repositories.RedisEventPublisher,
) *VoteService {
	return &VoteService{
		pollRepository: pollRepository,
		voteRepository: voteRepository,
		voteCounter:    voteCounter,
		eventPublisher: eventPublisher,
	}
}

func (s *VoteService) Vote(
	pollID string,
	optionID string,
	optionIndex *int,
	voterID string,
) (*models.PollResponseDTO, error) {
	pollID = strings.TrimSpace(pollID)
	optionID = strings.TrimSpace(optionID)
	voterID = strings.TrimSpace(voterID)

	if pollID == "" {
		return nil, errors.New("poll id is required")
	}

	poll, err := s.pollRepository.FindByID(pollID)
	if err != nil {
		return nil, errors.New("poll not found")
	}

	if poll.Status != "active" {
		return nil, errors.New("poll is closed")
	}

	targetIndex := -1

	if optionIndex != nil {
		idx := *optionIndex
		if idx < 0 || idx >= len(poll.Options) {
			return nil, errors.New("invalid option index")
		}
		targetIndex = idx
		optionID = poll.Options[idx].ID
	} else if optionID != "" {
		for i, option := range poll.Options {
			if option.ID == optionID {
				targetIndex = i
				break
			}
		}
		if targetIndex == -1 {
			return nil, errors.New("invalid poll option")
		}
	} else {
		return nil, errors.New("option index or option id is required")
	}

	// Persist vote in MongoDB
	vote := &models.Vote{
		ID:        uuid.NewString(),
		PollID:    poll.ID,
		OptionID:  optionID,
		VoterID:   voterID,
		CreatedAt: time.Now(),
	}

	if err := s.voteRepository.Create(vote); err != nil {
		return nil, err
	}

	ctx := context.Background()

	newCount, err := s.voteCounter.Increment(
		ctx,
		poll.ID,
		optionID,
	)
	if err != nil {
		return nil, errors.New("vote saved but live counter update failed")
	}

	votesInt64 := make([]int64, len(poll.Options))
	votesInt := make([]int, len(poll.Options))
	var totalVotes int64

	for i, option := range poll.Options {
		count, err := s.voteCounter.Get(
			ctx,
			poll.ID,
			option.ID,
		)
		if err == nil {
			votesInt64[i] = count
			votesInt[i] = int(count)
		}
		totalVotes += votesInt64[i]
	}

	percentages := make([]int, len(poll.Options))
	if totalVotes > 0 {
		sumPercent := 0
		for i, v := range votesInt64 {
			pct := int(math.Round(float64(v) / float64(totalVotes) * 100))
			percentages[i] = pct
			sumPercent += pct
		}
		if len(percentages) > 0 && sumPercent != 100 {
			percentages[0] += (100 - sumPercent)
		}
	}

	// Publish vote event to Redis Pub/Sub for realtime WebSocket broadcasting
	_ = s.eventPublisher.PublishVoteEvent(
		ctx,
		models.VoteEvent{
			PollID:      poll.ID,
			PollCode:    poll.Code,
			OptionID:    optionID,
			OptionIndex: targetIndex,
			Votes:       votesInt64,
			OptionVotes: newCount,
			TotalVotes:  totalVotes,
			Percentages: percentages,
		},
	)

	optionTexts := make([]string, len(poll.Options))
	for i, opt := range poll.Options {
		optionTexts[i] = opt.Text
	}

	code := poll.Code
	if code == "" {
		code = poll.ID
	}

	return &models.PollResponseDTO{
		ID:          poll.ID,
		Code:        code,
		Question:    poll.Question,
		Options:     optionTexts,
		RawOptions:  poll.Options,
		Votes:       votesInt,
		TotalVotes:  int(totalVotes),
		Percentages: percentages,
		Status:      poll.Status,
		CreatorID:   poll.CreatorID,
		CreatedAt:   poll.CreatedAt,
	}, nil
}

func (s *VoteService) GetResults(
	pollID string,
) ([]VoteResult, int, error) {
	pollID = strings.TrimSpace(pollID)

	if pollID == "" {
		return nil, 0, errors.New("poll id is required")
	}

	poll, err := s.pollRepository.FindByID(pollID)
	if err != nil {
		return nil, 0, errors.New("poll not found")
	}

	ctx := context.Background()

	results := make([]VoteResult, 0, len(poll.Options))
	totalVotes := 0

	for _, option := range poll.Options {
		count, err := s.voteCounter.Get(
			ctx,
			poll.ID,
			option.ID,
		)

		if err != nil {
			count = 0
		}

		results = append(results, VoteResult{
			OptionID:   option.ID,
			OptionText: option.Text,
			Votes:      int(count),
		})

		totalVotes += int(count)
	}

	for i := range results {
		if totalVotes > 0 {
			results[i].Percentage =
				(float64(results[i].Votes) /
					float64(totalVotes)) * 100
		}
	}

	return results, totalVotes, nil
}