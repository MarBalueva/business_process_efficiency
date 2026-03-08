package service

import (
	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/repository"
)

type MeasurementService struct {
	repo *repository.MeasurementRepository
}

func NewMeasurementService(repo *repository.MeasurementRepository) *MeasurementService {
	return &MeasurementService{repo: repo}
}

func (s *MeasurementService) StartMeasurement(stepID uint) (*models.StepMeasurement, error) {
	return s.repo.StartMeasurement(stepID)
}

func (s *MeasurementService) PauseMeasurement(id uint) error {
	return s.repo.PauseMeasurement(id)
}

func (s *MeasurementService) ResumeMeasurement(id uint) error {
	return s.repo.ResumeMeasurement(id)
}

func (s *MeasurementService) FinishMeasurement(id uint) error {
	return s.repo.FinishMeasurement(id)
}

func (s *MeasurementService) ResetMeasurements(stepID uint) error {
	return s.repo.ResetMeasurements(stepID)
}
