package repository

import (
	"gorm.io/gorm"
)

type DictionaryRepository struct {
	DB *gorm.DB
}

func NewDictionaryRepository(db *gorm.DB) *DictionaryRepository {
	return &DictionaryRepository{DB: db}
}

func (r *DictionaryRepository) List(model interface{}) error {
	return r.DB.Find(model).Error
}

func (r *DictionaryRepository) GetByID(model interface{}, id uint) error {
	return r.DB.First(model, id).Error
}

func (r *DictionaryRepository) Create(model interface{}) error {
	return r.DB.Create(model).Error
}

func (r *DictionaryRepository) Update(model interface{}) error {
	return r.DB.Save(model).Error
}

func (r *DictionaryRepository) Delete(model interface{}, id uint) error {
	return r.DB.Delete(model, id).Error
}
