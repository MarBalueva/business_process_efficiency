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
