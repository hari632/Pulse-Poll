package repositories

import (
	"sync"

	"pulse-backend/models"
)

type MemoryVoteRepository struct {
	mu    sync.RWMutex
	votes []*models.Vote
}

func NewMemoryVoteRepository() *MemoryVoteRepository {
	return &MemoryVoteRepository{
		votes: make([]*models.Vote, 0),
	}
}

func (r *MemoryVoteRepository) Create(vote *models.Vote) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.votes = append(r.votes, vote)

	return nil
}

func (r *MemoryVoteRepository) CountByPollAndOption(
	pollID string,
	optionID string,
) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0

	for _, vote := range r.votes {
		if vote.PollID == pollID && vote.OptionID == optionID {
			count++
		}
	}

	return count
}

func (r *MemoryVoteRepository) FindByPollID(
	pollID string,
) ([]*models.Vote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*models.Vote, 0)

	for _, vote := range r.votes {
		if vote.PollID == pollID {
			result = append(result, vote)
		}
	}

	return result, nil
}