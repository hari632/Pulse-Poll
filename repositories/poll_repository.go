package repositories

import "pulse-backend/models"

type PollRepository interface {
	Create(poll *models.Poll) error
	FindByID(id string) (*models.Poll, error)
	FindByCreatorID(creatorID string) ([]*models.Poll, error)
	Update(poll *models.Poll) error
}