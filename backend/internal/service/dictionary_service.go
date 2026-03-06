package service

import (
	"errors"

	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/repository"
)

type DictionaryService struct {
	repo *repository.DictionaryRepository
}

func NewDictionaryService(repo *repository.DictionaryRepository) *DictionaryService {
	return &DictionaryService{repo: repo}
}

func (s *DictionaryService) List(model interface{}) error {
	return s.repo.List(model)
}

func (s *DictionaryService) Create(model interface{}, input models.DictionaryRequest) error {

	switch m := model.(type) {

	case *models.Department:
		m.Name = input.Name
		m.Code = input.Code

	case *models.Position:
		m.Name = input.Name
		m.Code = input.Code

	case *models.AccessGroup:
		m.Name = input.Name
		m.Code = input.Code

	default:
		return errors.New("unknown dictionary type")
	}

	return s.repo.Create(model)
}

func (s *DictionaryService) Update(model interface{}, id uint, input models.DictionaryRequest) error {

	if err := s.repo.GetByID(model, id); err != nil {
		return err
	}

	switch m := model.(type) {

	case *models.Department:
		m.Name = input.Name
		m.Code = input.Code

	case *models.Position:
		m.Name = input.Name
		m.Code = input.Code

	case *models.AccessGroup:
		m.Name = input.Name
		m.Code = input.Code
	}

	return s.repo.Update(model)
}

func (s *DictionaryService) Delete(model interface{}, id uint) error {
	return s.repo.Delete(model, id)
}

func (s *DictionaryService) GetAll() (*models.DictionariesResponse, error) {

	var departments []models.Department
	var positions []models.Position
	var accessGroups []models.AccessGroup

	if err := s.repo.List(&departments); err != nil {
		return nil, err
	}

	if err := s.repo.List(&positions); err != nil {
		return nil, err
	}

	if err := s.repo.List(&accessGroups); err != nil {
		return nil, err
	}

	return &models.DictionariesResponse{
		Departments:  departments,
		Positions:    positions,
		AccessGroups: accessGroups,
	}, nil
}
