package service

import (
	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		repo: &repository.UserRepository{},
	}
}

func (s *UserService) CreateUser(user *models.User) error {
	return s.repo.Create(user)
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.repo.FindAll()
}
