package services

import (
	"context"
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	"pulse-backend/models"
	"pulse-backend/repositories"
)

const codeChars = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

func generatePollCode() string {
	b := make([]byte, 6)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeChars))))
		if err != nil {
			b[i] = codeChars[time.Now().UnixNano()%int64(len(codeChars))]
		} else {
			b[i] = codeChars[num.Int64()]
		}
	}
	return string(b)
}

type CreatePollInput struct {
	Question string
	Options  []string
}

type PollService struct {
	pollRepository repositories.PollRepository
	voteRepository repositories.VoteRepository
	voteCounter    *repositories.RedisVoteCounter
}

func NewPollService(
	pollRepository repositories.PollRepository,
	voteRepository repositories.VoteRepository,
	voteCounter *repositories.RedisVoteCounter,
) *PollService {
	return &PollService{
		pollRepository: pollRepository,
		voteRepository: voteRepository,
		voteCounter:    voteCounter,
	}
}

func (s *PollService) CreatePoll(
	creatorID string,
	input CreatePollInput,
) (*models.PollResponseDTO, error) {

	if strings.TrimSpace(creatorID) == "" {
		return nil, errors.New("creator is required")
	}

	question := strings.TrimSpace(input.Question)

	if question == "" {
		return nil, errors.New("question is required")
	}

	if len(question) < 5 {
		return nil, errors.New("question must be at least 5 characters")
	}

	if len(question) > 200 {
		return nil, errors.New("question must not exceed 200 characters")
	}

	if len(input.Options) < 2 {
		return nil, errors.New("poll must have at least 2 options")
	}

	if len(input.Options) > 10 {
		return nil, errors.New("poll cannot have more than 10 options")
	}

	options := make([]models.PollOption, 0, len(input.Options))
	seen := make(map[string]bool)

	for _, optionText := range input.Options {
		optionText = strings.TrimSpace(optionText)

		if optionText == "" {
			return nil, errors.New("options cannot be empty")
		}

		if len(optionText) > 100 {
			return nil, errors.New("option must not exceed 100 characters")
		}

		normalized := strings.ToLower(optionText)

		if seen[normalized] {
			return nil, errors.New("duplicate options are not allowed")
		}

		seen[normalized] = true

		options = append(options, models.PollOption{
			ID:   uuid.NewString(),
			Text: optionText,
		})
	}

	// Generate unique 6-character code for the poll
	var code string
	for i := 0; i < 5; i++ {
		c := generatePollCode()
		if _, err := s.pollRepository.FindByID(c); err != nil {
			code = c
			break
		}
	}
	if code == "" {
		code = generatePollCode()
	}

	now := time.Now()

	poll := &models.Poll{
		ID:              code,
		Code:            code,
		CreatorID:       creatorID,
		Question:        question,
		Options:         options,
		Status:          "active",
		AllowVoteChange: false,
		Anonymous:       false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.pollRepository.Create(poll); err != nil {
		return nil, err
	}

	return s.ToPollResponseDTO(context.Background(), poll), nil
}

func (s *PollService) GetPoll(id string) (*models.Poll, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("poll id is required")
	}

	return s.pollRepository.FindByID(id)
}

func (s *PollService) GetPollDTO(ctx context.Context, id string) (*models.PollResponseDTO, error) {
	poll, err := s.GetPoll(id)
	if err != nil {
		return nil, err
	}
	return s.ToPollResponseDTO(ctx, poll), nil
}

func (s *PollService) GetMyPolls(
	creatorID string,
) ([]*models.Poll, error) {

	if strings.TrimSpace(creatorID) == "" {
		return nil, errors.New("creator is required")
	}

	return s.pollRepository.FindByCreatorID(creatorID)
}

func (s *PollService) GetMyPollDTOs(ctx context.Context, creatorID string) ([]*models.PollResponseDTO, error) {
	polls, err := s.GetMyPolls(creatorID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*models.PollResponseDTO, len(polls))
	for i, poll := range polls {
		dtos[i] = s.ToPollResponseDTO(ctx, poll)
	}

	return dtos, nil
}

func (s *PollService) ClosePoll(userID, pollID string) (*models.PollResponseDTO, error) {
	poll, err := s.GetPoll(pollID)
	if err != nil {
		return nil, err
	}

	if poll.CreatorID != userID {
		return nil, errors.New("unauthorized to close this poll")
	}

	poll.Status = "closed"
	poll.UpdatedAt = time.Now()

	if err := s.pollRepository.Update(poll); err != nil {
		return nil, err
	}

	return s.ToPollResponseDTO(context.Background(), poll), nil
}

func (s *PollService) ToPollResponseDTO(ctx context.Context, poll *models.Poll) *models.PollResponseDTO {
	optionTexts := make([]string, len(poll.Options))
	votes := make([]int, len(poll.Options))
	totalVotes := 0

	for i, opt := range poll.Options {
		optionTexts[i] = opt.Text
		if s.voteCounter != nil {
			count, err := s.voteCounter.Get(ctx, poll.ID, opt.ID)
			if err == nil {
				votes[i] = int(count)
			}
		}
		totalVotes += votes[i]
	}

	percentages := make([]int, len(poll.Options))
	if totalVotes > 0 {
		sumPercent := 0
		for i, v := range votes {
			pct := int(math.Round(float64(v) / float64(totalVotes) * 100))
			percentages[i] = pct
			sumPercent += pct
		}
		if len(percentages) > 0 && sumPercent != 100 {
			percentages[0] += (100 - sumPercent)
		}
	}

	code := poll.Code
	if code == "" {
		code = poll.ID
	}

	peakActivity := "No activity yet"
	activity := "No votes yet"
	if s.voteRepository != nil {
		allVotes, err := s.voteRepository.FindByPollID(poll.ID)
		if err == nil {
			peakActivity, activity = calculateActivity(allVotes)
		}
	}

	return &models.PollResponseDTO{
		ID:           poll.ID,
		Code:         code,
		Question:     poll.Question,
		Options:      optionTexts,
		RawOptions:   poll.Options,
		Votes:        votes,
		TotalVotes:   totalVotes,
		Percentages:  percentages,
		Status:       poll.Status,
		PeakActivity: peakActivity,
		Activity:     activity,
		CreatorID:    poll.CreatorID,
		CreatedAt:    poll.CreatedAt,
	}
}