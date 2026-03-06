package service

import (
	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/repository"
	"time"
)

type EmployeeService struct {
	repo *repository.EmployeeRepository
}

func NewEmployeeService(repo *repository.EmployeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) Create(req models.EmployeeCreateRequest) error {
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return err
	}

	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		return err
	}

	employee := &models.Employee{
		LastName:   req.LastName,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		Code:       time.Now().Format("20060102150405"),
	}

	hr := &models.EmployeeHR{
		DepartmentID: req.DepartmentID,
		PositionID:   req.PositionID,
		IsRemote:     req.IsRemote,
		BirthDate:    birthDate,
		HireDate:     hireDate,
		Salary:       req.Salary,
	}

	return s.repo.Create(employee, hr)
}

func (s *EmployeeService) GetAll() ([]models.EmployeeFullResponse, error) {
	return s.repo.GetAll()
}

func (s *EmployeeService) GetByID(id uint) (*models.EmployeeFullResponse, error) {
	return s.repo.GetByID(id)
}

func (s *EmployeeService) Update(id uint, req models.EmployeeCreateRequest) error {
	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return err
	}

	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		return err
	}

	return s.repo.Update(id, req, birthDate, hireDate)
}

func (s *EmployeeService) Delete(id uint) error {
	return s.repo.Delete(id)
}
