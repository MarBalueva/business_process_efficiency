package models

import (
	"time"

	"gorm.io/gorm"
)

type Employee struct {
	ID        uint   `gorm:"primaryKey"`
	Code      string `gorm:"size:255;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" swaggerignore:"true"`

	LastName   string `gorm:"size:255;not null"`
	FirstName  string `gorm:"size:255;not null"`
	MiddleName string `gorm:"size:255"`
}

type EmployeeHR struct {
	ID uint `gorm:"primaryKey"`

	DepartmentID uint
	Department   Department `gorm:"foreignKey:DepartmentID"`

	PositionID uint
	Position   Position `gorm:"foreignKey:PositionID"`

	EmployeeID uint
	Employee   Employee `gorm:"foreignKey:EmployeeID"`

	IsRemote  bool
	BirthDate time.Time
	HireDate  time.Time
	FireDate  *time.Time
	Salary    float64

	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" swaggerignore:"true"`
}

type ProfileUpdateRequest struct {
	Password string `json:"password" example:"secret123"`

	LastName   string `json:"last_name" example:"Иванов"`
	FirstName  string `json:"first_name" example:"Иван"`
	MiddleName string `json:"middle_name" example:"Иванович"`
}

type EmployeeFullResponse struct {
	ID uint `json:"id"`

	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`

	Department string `json:"department"`
	Position   string `json:"position"`

	IsRemote  bool       `json:"is_remote"`
	BirthDate time.Time  `json:"birth_date"`
	HireDate  time.Time  `json:"hire_date"`
	FireDate  *time.Time `json:"fire_date"`
	Salary    float64    `json:"salary"`

	CreatedAt time.Time `json:"created_at"`
}

type EmployeeCreateRequest struct {
	LastName   string `json:"last_name" binding:"required"`
	FirstName  string `json:"first_name" binding:"required"`
	MiddleName string `json:"middle_name"`

	DepartmentID uint    `json:"department_id" binding:"required"`
	PositionID   uint    `json:"position_id" binding:"required"`
	IsRemote     bool    `json:"is_remote"`
	BirthDate    string  `json:"birth_date" binding:"required"`
	HireDate     string  `json:"hire_date" binding:"required"`
	Salary       float64 `json:"salary"`
}
