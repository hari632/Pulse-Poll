package repositories

import (
	"errors"
	"sync"

	"pulse-backend/models"
)

type MemoryPollRepository struct {
	mu    sync.RWMutex
	polls map[string]*models.Poll
}

func NewMemoryPollRepository() *MemoryPollRepository {
	return &MemoryPollRepository{
		polls: make(map[string]*models.Poll),
	}
}

func (r *MemoryPollRepository) Create(poll *models.Poll) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.polls[poll.ID] = poll

	return nil
}

func (r *MemoryPollRepository) FindByID(id string) (*models.Poll, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	poll, exists := r.polls[id]

	if !exists {
		return nil, errors.New("poll not found")
	}

	return poll, nil
}

func (r *MemoryPollRepository) FindByCreatorID(creatorID string) ([]*models.Poll, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Poll

	for _, poll := range r.polls {
		if poll.CreatorID == creatorID {
			result = append(result, poll)
		}
	}

	return result, nil
}

func (r *MemoryPollRepository) Update(poll *models.Poll) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.polls[poll.ID]; !exists {
		return errors.New("poll not found")
	}

	r.polls[poll.ID] = poll

	return nil
}