package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"pulse-backend/models"
	"pulse-backend/repositories"
	"pulse-backend/utils"
)

type AuthService struct {
	userRepository repositories.UserRepository
}

func NewAuthService(userRepository repositories.UserRepository) *AuthService {
	return &AuthService{
		userRepository: userRepository,
	}
}

func (s *AuthService) Register(
	name string,
	email string,
	password string,
) (*models.User, error) {

	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))

	if name == "" {
		return nil, errors.New("name is required")
	}

	if email == "" {
		return nil, errors.New("email is required")
	}

	if password == "" {
		return nil, errors.New("password is required")
	}

	if len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	_, err := s.userRepository.FindByEmail(email)

	if err == nil {
		return nil, errors.New("email already exists")
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	now := time.Now()

	user := &models.User{
		ID:           uuid.NewString(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepository.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(
	email string,
	password string,
) (string, *models.User, error) {

	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" || password == "" {
		return "", nil, errors.New("email and password are required")
	}

	user, err := s.userRepository.FindByEmail(email)

	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return "", nil, errors.New("invalid email or password")
	}

	token, err := utils.GenerateToken(user.ID, user.Email)

	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	return token, user, nil
}
func (s *AuthService) GetUserByID(id string) (*models.User, error) {
	return s.userRepository.FindByID(id)
}