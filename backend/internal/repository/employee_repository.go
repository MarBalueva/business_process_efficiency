package repository

import (
	"business_process_efficiency/internal/database"
	"business_process_efficiency/internal/models"
	"time"

	"gorm.io/gorm"
)

type EmployeeRepository struct {
	DB *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{DB: db}
}

func (r *EmployeeRepository) Create(employee *models.Employee, hr *models.EmployeeHR) error {

	tx := database.DB.Begin()

	if err := tx.Create(employee).Error; err != nil {
		tx.Rollback()
		return err
	}

	hr.EmployeeID = employee.ID

	if err := tx.Create(hr).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *EmployeeRepository) CreateHR(hr *models.EmployeeHR) error {
	return r.DB.Create(hr).Error
}

func (r *EmployeeRepository) GetAll() ([]models.EmployeeFullResponse, error) {

	var employees []models.EmployeeFullResponse

	err := r.DB.
		Table("employees").
		Select(`
			employees.id,
			employees.last_name,
			employees.first_name,
			employees.middle_name,
			departments.name as department,
			positions.name as position,
			employee_hrs.is_remote,
			employee_hrs.birth_date,
			employee_hrs.hire_date,
			employee_hrs.fire_date,
			employee_hrs.salary,
			employees.created_at
		`).
		Joins("left join employee_hrs on employee_hrs.employee_id = employees.id").
		Joins("left join departments on departments.id = employee_hrs.department_id").
		Joins("left join positions on positions.id = employee_hrs.position_id").
		Where("employees.deleted_at IS NULL").
		Scan(&employees).Error

	return employees, err
}

func (r *EmployeeRepository) GetByID(id uint) (*models.EmployeeFullResponse, error) {

	var hr models.EmployeeHR

	err := database.DB.
		Preload("Employee").
		Preload("Department").
		Preload("Position").
		Where("employee_id = ?", id).
		First(&hr).Error

	if err != nil {
		return nil, err
	}

	resp := models.EmployeeFullResponse{
		ID:         hr.Employee.ID,
		LastName:   hr.Employee.LastName,
		FirstName:  hr.Employee.FirstName,
		MiddleName: hr.Employee.MiddleName,
		Department: hr.Department.Name,
		Position:   hr.Position.Name,
		IsRemote:   hr.IsRemote,
		BirthDate:  hr.BirthDate,
		HireDate:   hr.HireDate,
		FireDate:   hr.FireDate,
		CreatedAt:  hr.Employee.CreatedAt,
		Salary:     hr.Salary,
	}

	return &resp, nil
}

func (r *EmployeeRepository) Update(
	id uint,
	req models.EmployeeCreateRequest,
	birthDate time.Time,
	hireDate time.Time,
) error {

	tx := database.DB.Begin()

	err := tx.Model(&models.Employee{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_name":   req.LastName,
			"first_name":  req.FirstName,
			"middle_name": req.MiddleName,
		}).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.EmployeeHR{}).
		Where("employee_id = ?", id).
		Updates(map[string]interface{}{
			"department_id": req.DepartmentID,
			"position_id":   req.PositionID,
			"is_remote":     req.IsRemote,
			"birth_date":    birthDate,
			"hire_date":     hireDate,
			"salary":        req.Salary,
		}).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *EmployeeRepository) Delete(id uint) error {

	tx := database.DB.Begin()

	if err := tx.Where("employee_id = ?", id).
		Delete(&models.EmployeeHR{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&models.Employee{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
