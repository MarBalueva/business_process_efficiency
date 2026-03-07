package service

import (
	"business_process_efficiency/internal/database"
	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ProfileService struct {
	repo *repository.ProfileRepository
}

func NewProfileService(repo *repository.ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

// Получение профиля текущего пользователя
func (s *ProfileService) GetProfile(userID uint) (*models.User, *models.Employee, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, nil, err
	}

	employee, err := s.repo.GetEmployeeByID(user.EmployeeID)
	if err != nil {
		return nil, nil, err
	}

	return user, employee, nil
}

func (s *ProfileService) UpdateProfile(userID uint, input models.ProfileUpdateRequest) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return err
	}

	employee, err := s.repo.GetEmployeeByID(user.EmployeeID)
	if err != nil {
		return err
	}

	if input.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashed)
	}

	if input.LastName != "" {
		employee.LastName = input.LastName
	}
	if input.FirstName != "" {
		employee.FirstName = input.FirstName
	}
	if input.MiddleName != "" {
		employee.MiddleName = input.MiddleName
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.UpdateUser(user, tx); err != nil {
			return err
		}
		if err := s.repo.UpdateEmployee(employee, tx); err != nil {
			return err
		}
		return nil
	})
}
