package models

import (
	"time"
)

type User struct {
	ID         uint       `gorm:"primaryKey;column:id"`
	Login      string     `gorm:"column:login;unique;not null"`
	Password   string     `gorm:"column:password;not null"`
	EmployeeID uint       `gorm:"column:employee_id;not null"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
	DeletedAt  *time.Time `gorm:"column:deleted_at" json:"-" swaggerignore:"true"`

	AccessGroups []UserAccessGroup `gorm:"foreignKey:UserID"`
}

type ProfileUpdateRequest struct {
	Password string `json:"password" example:"secret123"`

	LastName   string `json:"last_name" example:"Иванов"`
	FirstName  string `json:"first_name" example:"Иван"`
	MiddleName string `json:"middle_name" example:"Иванович"`
}

type UserAccessGroup struct {
	UserID uint

	User User `gorm:"foreignKey:UserID" json:"-" swaggerignore:"true"`

	AccessGroupID uint

	AccessGroup AccessGroup `gorm:"foreignKey:AccessGroupID" json:"-" swaggerignore:"true"`

	CreatedAt time.Time
	DeletedAt *time.Time `gorm:"index" json:"-" swaggerignore:"true"`
}

type CreateUserRequest struct {
	Login      string `json:"login" binding:"required"`
	Password   string `json:"password" binding:"required"`
	EmployeeID uint   `json:"employee_id" binding:"required"`
}

type UpdateUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
