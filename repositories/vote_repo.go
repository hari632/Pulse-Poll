package repositories

import "pulse-backend/models"

type VoteRepository interface {
	Create(vote *models.Vote) error
	CountByPollAndOption(pollID, optionID string) int
	FindByPollID(pollID string) ([]*models.Vote, error)
}