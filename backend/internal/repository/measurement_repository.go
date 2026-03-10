package repository

import (
	"business_process_efficiency/internal/models"
	"errors"
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
	var metricsCount int64
	if err := r.db.Model(&models.StepMetrics{}).Where("step_id = ?", stepID).Count(&metricsCount).Error; err != nil {
		return nil, err
	}
	if metricsCount > 0 {
		return nil, errors.New("cannot start measurement: step has statistical data")
	}

	var count int64
	if err := r.db.Model(&models.StepMeasurement{}).Where("step_id = ?", stepID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count >= 3 {
		return nil, errors.New("maximum 3 measurements per step")
	}

	now := time.Now()
	measurementNumber := int(count) + 1

	// Use explicit map insert so numeric fields are written deterministically.
	if err := r.db.Model(&models.StepMeasurement{}).Create(map[string]interface{}{
		"step_id":            stepID,
		"measurement_number": measurementNumber,
		"started_at":         now,
		"paused_seconds":     0,
		"duration_sec":       0,
		"finished_at":        nil,
	}).Error; err != nil {
		return nil, err
	}

	var m models.StepMeasurement
	if err := r.db.Where("step_id = ?", stepID).Order("id DESC").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MeasurementRepository) PauseMeasurement(id uint) error {
	var openPause models.MeasurementPause
	err := r.db.Where("measurement_id = ? AND pause_end IS NULL", id).First(&openPause).Error
	if err == nil {
		return errors.New("measurement is already paused")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

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
		Where("measurement_id = ? AND pause_end IS NULL", id).
		Last(&pause).Error

	if err != nil {
		return err
	}

	now := time.Now()
	pause.PauseEnd = &now

	if err := r.db.Save(&pause).Error; err != nil {
		return err
	}

	paused := int(now.Sub(pause.PauseStart).Seconds())
	if paused < 0 {
		paused = 0
	}

	return r.db.Model(&models.StepMeasurement{}).
		Where("id = ?", id).
		Update("paused_seconds", gorm.Expr("paused_seconds + ?", paused)).Error
}

func (r *MeasurementRepository) FinishMeasurement(id uint) error {

	var m models.StepMeasurement

	err := r.db.First(&m, id).Error
	if err != nil {
		return err
	}

	now := time.Now()
	totalPaused := m.PausedSeconds

	// If measurement is currently paused, close the pause and count it.
	var openPause models.MeasurementPause
	err = r.db.
		Where("measurement_id = ? AND pause_end IS NULL", id).
		Last(&openPause).Error
	if err == nil {
		openPause.PauseEnd = &now
		if err := r.db.Save(&openPause).Error; err != nil {
			return err
		}

		paused := int(now.Sub(openPause.PauseStart).Seconds())
		if paused > 0 {
			totalPaused += paused
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	duration := int(now.Sub(*m.StartedAt).Seconds()) - totalPaused
	if duration < 0 {
		duration = 0
	}

	if err := r.db.Model(&models.StepMeasurement{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"finished_at":    now,
			"paused_seconds": totalPaused,
			"duration_sec":   duration,
		}).Error; err != nil {
		return err
	}

	return r.recalculateStepFinalDuration(m.StepID)
}

func (r *MeasurementRepository) ResetMeasurements(stepID uint) error {
	var measurementIDs []uint
	if err := r.db.Model(&models.StepMeasurement{}).Where("step_id = ?", stepID).Pluck("id", &measurementIDs).Error; err != nil {
		return err
	}

	if len(measurementIDs) > 0 {
		if err := r.db.Where("measurement_id IN ?", measurementIDs).Delete(&models.MeasurementPause{}).Error; err != nil {
			return err
		}
	}

	if err := r.db.Where("step_id = ?", stepID).Delete(&models.StepMeasurement{}).Error; err != nil {
		return err
	}

	return r.recalculateStepFinalDuration(stepID)
}

func (r *MeasurementRepository) DeleteMeasurement(id uint) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var measurement models.StepMeasurement
	if err := tx.Select("id", "step_id").First(&measurement, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("measurement_id = ?", id).Delete(&models.MeasurementPause{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&models.StepMeasurement{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := r.recalculateStepFinalDurationTx(tx, measurement.StepID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *MeasurementRepository) GetMeasurementsByStep(stepID uint) ([]models.StepMeasurement, error) {
	var list []models.StepMeasurement
	err := r.db.
		Where("step_id = ?", stepID).
		Preload("Pauses").
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *MeasurementRepository) recalculateStepFinalDuration(stepID uint) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := r.recalculateStepFinalDurationTx(tx, stepID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *MeasurementRepository) recalculateStepFinalDurationTx(tx *gorm.DB, stepID uint) error {
	var finishedCount int64
	if err := tx.Model(&models.StepMeasurement{}).
		Where("step_id = ? AND finished_at IS NOT NULL", stepID).
		Count(&finishedCount).Error; err != nil {
		return err
	}

	finalDurationMin := 0.0

	if finishedCount > 0 {
		var avgDurationSec float64
		if err := tx.Model(&models.StepMeasurement{}).
			Where("step_id = ? AND finished_at IS NOT NULL", stepID).
			Select("COALESCE(AVG(duration_sec), 0)").
			Scan(&avgDurationSec).Error; err != nil {
			return err
		}
		finalDurationMin = avgDurationSec / 60.0
	} else {
		var metrics models.StepMetrics
		err := tx.Preload("TimeStatistics").Where("step_id = ?", stepID).First(&metrics).Error
		if err == nil && metrics.TimeStatistics != nil {
			finalDurationMin = metrics.TimeStatistics.WeightedAvg
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	return tx.Model(&models.ProcessStep{}).
		Where("id = ?", stepID).
		Update("final_duration_min", finalDurationMin).Error
}
