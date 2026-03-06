package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseDictionary struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:255;not null"`
	Code      string `gorm:"column:code;unique"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" swaggerignore:"true"`
}

type Department struct {
	BaseDictionary
}

type Position struct {
	BaseDictionary
}

type AccessGroup struct {
	BaseDictionary
}

type DictionaryRequest struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code"`
}

type DictionaryModel interface {
	TableName() string
}

type DepartmentResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type PositionResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type AccessGroupResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type DictionariesResponse struct {
	Departments  []Department  `json:"departments"`
	Positions    []Position    `json:"positions"`
	AccessGroups []AccessGroup `json:"access_groups"`
}
