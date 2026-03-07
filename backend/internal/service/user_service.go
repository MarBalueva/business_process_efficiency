package service

import (
	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/repository"

	"golang.org/x/crypto/bcrypt"
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

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hash)

	return s.repo.Create(user)
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.repo.FindAll()
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) UpdateUser(id uint, login string, password string) error {

	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if login != "" {
		user.Login = login
	}

	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hash)
	}

	return s.repo.Update(user)
}

func (s *UserService) DeleteUser(id uint) error {
	return s.repo.SoftDelete(id)
}

func (s *UserService) AddAccessGroup(userID uint, groupID uint) error {
	return s.repo.AddAccessGroup(userID, groupID)
}

func (s *UserService) RemoveAccessGroup(userID uint, groupID uint) error {
	return s.repo.RemoveAccessGroup(userID, groupID)
}

func (s *UserService) GetUserAccessGroups(userID uint) ([]models.UserAccessGroup, error) {
	return s.repo.GetUserAccessGroups(userID)
}

func (s *UserService) GetUserByEmployeeID(employeeID uint) (*models.User, error) {
	return s.repo.GetByEmployeeID(employeeID)
}
