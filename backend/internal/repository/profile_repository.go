package repository

import (
	"business_process_efficiency/internal/models"

	"gorm.io/gorm"
)

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *ProfileRepository) GetEmployeeByID(employeeID uint) (*models.Employee, error) {
	var emp models.Employee
	if err := r.db.First(&emp, employeeID).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *ProfileRepository) UpdateUser(user *models.User, tx *gorm.DB) error {
	return tx.Save(user).Error
}

func (r *ProfileRepository) UpdateEmployee(emp *models.Employee, tx *gorm.DB) error {
	return tx.Save(emp).Error
}
