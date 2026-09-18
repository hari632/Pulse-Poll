package repositories

import (
	"errors"
	"sync"

	"pulse-backend/models"
)

type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*models.User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users: make(map[string]*models.User),
	}
}

func (r *MemoryUserRepository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.users {
		if existing.Email == user.Email {
			return errors.New("email already exists")
		}
	}

	r.users[user.ID] = user

	return nil
}

func (r *MemoryUserRepository) FindByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

func (r *MemoryUserRepository) FindByID(id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]

	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}