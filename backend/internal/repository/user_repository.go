package repository

import (
	"time"

	"business_process_efficiency/internal/database"
	"business_process_efficiency/internal/models"
)

type UserRepository struct{}

func (r *UserRepository) Create(user *models.User) error {
	return database.DB.Create(user).Error
}

func (r *UserRepository) FindAll() ([]models.User, error) {
	var users []models.User

	err := database.DB.
		Preload("AccessGroups", "deleted_at IS NULL").
		Preload("AccessGroups.AccessGroup", "deleted_at IS NULL").
		Where("deleted_at IS NULL").
		Find(&users).Error

	return users, err
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User

	err := database.DB.
		Preload("AccessGroups", "deleted_at IS NULL").
		Preload("AccessGroups.AccessGroup", "deleted_at IS NULL").
		Where("deleted_at IS NULL").
		First(&user, id).Error

	return &user, err
}

func (r *UserRepository) Update(user *models.User) error {
	return database.DB.Save(user).Error
}

func (r *UserRepository) SoftDelete(id uint) error {
	now := time.Now()

	return database.DB.
		Model(&models.User{}).
		Where("id = ?", id).
		Update("deleted_at", &now).Error
}

func (r *UserRepository) AddAccessGroup(userID uint, groupID uint) error {
	access := models.UserAccessGroup{
		UserID:        userID,
		AccessGroupID: groupID,
	}

	return database.DB.Create(&access).Error
}

func (r *UserRepository) RemoveAccessGroup(userID uint, groupID uint) error {
	now := time.Now()

	return database.DB.
		Model(&models.UserAccessGroup{}).
		Where("user_id = ? AND access_group_id = ? AND deleted_at IS NULL", userID, groupID).
		Update("deleted_at", &now).Error
}

func (r *UserRepository) GetUserAccessGroups(userID uint) ([]models.UserAccessGroup, error) {
	var groups []models.UserAccessGroup

	err := database.DB.
		Preload("AccessGroup", "deleted_at IS NULL").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Find(&groups).Error

	return groups, err
}

func (r *UserRepository) GetByEmployeeID(employeeID uint) (*models.User, error) {
	var user models.User

	err := database.DB.
		Where("employee_id = ? AND deleted_at IS NULL", employeeID).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}
