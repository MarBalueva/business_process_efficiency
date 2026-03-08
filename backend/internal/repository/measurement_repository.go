package repository

import (
	"business_process_efficiency/internal/models"
	"time"

	"gorm.io/gorm"
)

type MeasurementRepository struct {
	db *gorm.DB
}

func NewMeasurementRepository(db *gorm.DB) *MeasurementRepository {
	return &MeasurementRepository{db: db}
}

func (r *MeasurementRepository) StartMeasurement(stepID uint) (*models.StepMeasurement, error) {

	now := time.Now()

	m := models.StepMeasurement{
		StepID:    stepID,
		StartedAt: &now,
	}

	err := r.db.Create(&m).Error
	return &m, err
}

func (r *MeasurementRepository) PauseMeasurement(id uint) error {

	now := time.Now()

	pause := models.MeasurementPause{
		MeasurementID: id,
		PauseStart:    now,
	}

	return r.db.Create(&pause).Error
}

func (r *MeasurementRepository) ResumeMeasurement(id uint) error {

	var pause models.MeasurementPause

	err := r.db.
		Where("measurement_id = ? AND resumed_at IS NULL", id).
		Last(&pause).Error

	if err != nil {
		return err
	}

	now := time.Now()
	pause.PauseEnd = &now

	return r.db.Save(&pause).Error
}

func (r *MeasurementRepository) FinishMeasurement(id uint) error {

	var m models.StepMeasurement

	err := r.db.First(&m, id).Error
	if err != nil {
		return err
	}

	now := time.Now()

	m.FinishedAt = &now
	duration := int(now.Sub(*m.StartedAt).Seconds())

	m.DurationSec = duration

	return r.db.Save(&m).Error
}

func (r *MeasurementRepository) ResetMeasurements(stepID uint) error {

	err := r.db.
		Where("step_id = ?", stepID).
		Delete(&models.StepMeasurement{}).Error

	if err != nil {
		return err
	}

	return r.db.
		Where("measurement_id IN (?)",
			r.db.Model(&models.StepMeasurement{}).
				Select("id").
				Where("step_id = ?", stepID),
		).
		Delete(&models.MeasurementPause{}).Error
}
