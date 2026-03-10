package repository

import (
	"business_process_efficiency/internal/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type ProcessRepository struct {
	db *gorm.DB
}

func NewProcessRepository(db *gorm.DB) *ProcessRepository {
	return &ProcessRepository{db: db}
}

func (r *ProcessRepository) GetRegistry() ([]models.ProcessFolder, error) {
	var folders []models.ProcessFolder

	err := r.db.
		Preload("Processes").
		Find(&folders).Error

	return folders, err
}

func (r *ProcessRepository) GetProcessByID(id uint) (*models.Process, error) {
	var process models.Process

	err := r.db.
		Preload("Owner").
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("version ASC")
		}).
		Preload("Versions.Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order ASC, id ASC")
		}).
		Preload("Versions.Steps.Executors").
		Preload("Versions.Steps.StepExecutors.Employee").
		Preload("Versions.Steps.Metrics.TimeStatistics").
		Preload("Versions.Steps.Measurements").
		Preload("Versions.Steps.Measurements.Pauses").
		First(&process, id).Error

	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (r *ProcessRepository) GetStepByID(id uint) (*models.ProcessStep, error) {
	var step models.ProcessStep

	err := r.db.
		Preload("StepExecutors.Employee").
		Preload("Metrics.TimeStatistics").
		Preload("Measurements").
		Preload("Measurements.Pauses").
		First(&step, id).Error
	if err != nil {
		return nil, err
	}

	return &step, nil
}

func (r *ProcessRepository) CreateProcess(process *models.Process) error {
	return r.db.Create(process).Error
}

func (r *ProcessRepository) UpdateProcess(id uint, req models.UpdateProcessRequest) (*models.Process, error) {

	var process models.Process

	err := r.db.First(&process, id).Error
	if err != nil {
		return nil, err
	}

	err = r.db.Model(&process).Updates(map[string]interface{}{
		"name":             req.Name,
		"description":      req.Description,
		"regulations":      req.Regulations,
		"owner_id":         req.OwnerID,
		"is_active":        req.IsActive,
		"regularity_count": req.RegularityCount,
		"regularity_unit":  req.RegularityUnit,
	}).Error

	if err != nil {
		return nil, err
	}

	err = r.db.Preload("Owner").First(&process, id).Error
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (r *ProcessRepository) DeleteProcess(id uint) error {
	return r.db.Delete(&models.Process{}, id).Error
}

func (r *ProcessRepository) CreateVersion(version *models.ProcessVersion) error {
	return r.db.Create(version).Error
}

func (r *ProcessRepository) GetLastVersionNumber(processID uint) (int, error) {

	var version models.ProcessVersion

	err := r.db.
		Where("process_id = ?", processID).
		Order("version DESC").
		First(&version).Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}

		return 0, err
	}

	return version.Version, nil
}

func (r *ProcessRepository) DeleteVersion(id uint) error {
	return r.db.Delete(&models.ProcessVersion{}, id).Error
}

func (r *ProcessRepository) CreateStep(step *models.ProcessStep) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	base := models.ProcessStep{
		ProcessVersionID: step.ProcessVersionID,
		StepOrder:        step.StepOrder,
		Name:             step.Name,
		Type:             step.Type,
		Description:      step.Description,
		SubprocessID:     step.SubprocessID,
		FinalDurationMin: step.FinalDurationMin,
	}

	if err := tx.Create(&base).Error; err != nil {
		tx.Rollback()
		return err
	}
	step.ID = base.ID

	if err := r.replaceStepExecutorsTx(tx, step.ID, step.StepExecutors, step.Executors); err != nil {
		tx.Rollback()
		return err
	}

	if err := r.rebuildStepLinksTx(tx, step.ProcessVersionID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *ProcessRepository) GetLastStepByVersion(versionID uint, step *models.ProcessStep) error {
	return r.db.Where("process_version_id = ?", versionID).Order("step_order desc").First(step).Error
}

func (r *ProcessRepository) GetEmployeesByIDs(ids []uint, employees *[]models.Employee) error {
	return r.db.Where("id IN ?", ids).Find(employees).Error
}

func (r *ProcessRepository) UpdateStep(step *models.ProcessStep) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var currentStep models.ProcessStep
	if err := tx.Select("id", "process_version_id").First(&currentStep, step.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	updateFields := map[string]interface{}{
		"name":        step.Name,
		"type":        step.Type,
		"description": step.Description,
	}

	if err := tx.Model(&models.ProcessStep{}).Where("id = ?", step.ID).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Replace only join table rows and persist workload percentages.
	if step.StepExecutors != nil || step.Executors != nil {
		if err := r.replaceStepExecutorsTx(tx, step.ID, step.StepExecutors, step.Executors); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Metrics toggle: nil means "disable metrics" from frontend.
	if step.Metrics == nil {
		var existing models.StepMetrics
		if err := tx.Where("step_id = ?", step.ID).First(&existing).Error; err == nil {
			if err := tx.Where("metrics_id = ?", existing.ID).Delete(&models.StepTimeStatistics{}).Error; err != nil {
				tx.Rollback()
				return err
			}
			if err := tx.Delete(&models.StepMetrics{}, existing.ID).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return err
		}

		if err := r.recalculateStepFinalDurationTx(tx, step.ID); err != nil {
			tx.Rollback()
			return err
		}

		if err := r.rebuildStepLinksTx(tx, currentStep.ProcessVersionID); err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit().Error
	}

	var metric models.StepMetrics
	err := tx.Where("step_id = ?", step.ID).First(&metric).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		metric = models.StepMetrics{
			StepID:         step.ID,
			PlannedTimeMin: step.Metrics.PlannedTimeMin,
		}
		if err := tx.Create(&metric).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else if err != nil {
		tx.Rollback()
		return err
	} else {
		if err := tx.Model(&models.StepMetrics{}).Where("id = ?", metric.ID).
			Update("planned_time_min", step.Metrics.PlannedTimeMin).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if step.Metrics.TimeStatistics != nil {
		var measurementsCount int64
		if err := tx.Model(&models.StepMeasurement{}).Where("step_id = ?", step.ID).Count(&measurementsCount).Error; err != nil {
			tx.Rollback()
			return err
		}
		if measurementsCount > 0 {
			tx.Rollback()
			return errors.New("cannot set statistical data: step already has measurements")
		}

		ts := step.Metrics.TimeStatistics
		var existingTS models.StepTimeStatistics
		err := tx.Where("metrics_id = ?", metric.ID).First(&existingTS).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newTS := models.StepTimeStatistics{
				MetricsID:   metric.ID,
				MinTime:     ts.MinTime,
				MinPercent:  ts.MinPercent,
				AvgTime:     ts.AvgTime,
				AvgPercent:  ts.AvgPercent,
				MaxTime:     ts.MaxTime,
				MaxPercent:  ts.MaxPercent,
				WeightedAvg: ts.WeightedAvg,
			}
			if err := tx.Create(&newTS).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else if err != nil {
			tx.Rollback()
			return err
		} else {
			updates := map[string]interface{}{
				"min_time":     ts.MinTime,
				"min_percent":  ts.MinPercent,
				"avg_time":     ts.AvgTime,
				"avg_percent":  ts.AvgPercent,
				"max_time":     ts.MaxTime,
				"max_percent":  ts.MaxPercent,
				"weighted_avg": ts.WeightedAvg,
			}
			if err := tx.Model(&models.StepTimeStatistics{}).Where("id = ?", existingTS.ID).Updates(updates).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	} else {
		// Statistics are disabled, keep planned time but remove statistical data.
		if err := tx.Where("metrics_id = ?", metric.ID).Delete(&models.StepTimeStatistics{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := r.recalculateStepFinalDurationTx(tx, step.ID); err != nil {
		tx.Rollback()
		return err
	}

	if err := r.rebuildStepLinksTx(tx, currentStep.ProcessVersionID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *ProcessRepository) recalculateStepFinalDurationTx(tx *gorm.DB, stepID uint) error {
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

func (r *ProcessRepository) DeleteStep(id uint) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var step models.ProcessStep
	if err := tx.Select("id", "process_version_id").First(&step, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("from_step_id = ? OR to_step_id = ?", id, id).Delete(&models.ProcessStepLink{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&models.ProcessStep{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	var steps []models.ProcessStep
	if err := tx.
		Where("process_version_id = ?", step.ProcessVersionID).
		Order("step_order ASC, id ASC").
		Find(&steps).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i, s := range steps {
		expectedOrder := i + 1
		if s.StepOrder == expectedOrder {
			continue
		}

		if err := tx.Model(&models.ProcessStep{}).
			Where("id = ?", s.ID).
			Update("step_order", expectedOrder).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := r.rebuildStepLinksTx(tx, step.ProcessVersionID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *ProcessRepository) rebuildStepLinksTx(tx *gorm.DB, processVersionID uint) error {
	var steps []models.ProcessStep
	if err := tx.
		Select("id").
		Where("process_version_id = ?", processVersionID).
		Order("step_order ASC, id ASC").
		Find(&steps).Error; err != nil {
		return err
	}

	ids := make([]uint, 0, len(steps))
	for _, s := range steps {
		ids = append(ids, s.ID)
	}

	if len(ids) > 0 {
		if err := tx.Where("from_step_id IN ? OR to_step_id IN ?", ids, ids).Delete(&models.ProcessStepLink{}).Error; err != nil {
			return err
		}
	} else {
		return nil
	}

	if len(steps) < 2 {
		return nil
	}

	links := make([]models.ProcessStepLink, 0, len(steps)-1)
	for i := 0; i < len(steps)-1; i++ {
		links = append(links, models.ProcessStepLink{
			FromStepID: steps[i].ID,
			ToStepID:   steps[i+1].ID,
		})
	}

	return tx.Create(&links).Error
}

func (r *ProcessRepository) replaceStepExecutorsTx(
	tx *gorm.DB,
	stepID uint,
	stepExecutors []models.ProcessStepExecutor,
	executors []models.Employee,
) error {
	if err := tx.Where("process_step_id = ?", stepID).Delete(&models.ProcessStepExecutor{}).Error; err != nil {
		return err
	}

	rows := make([]models.ProcessStepExecutor, 0, len(stepExecutors)+len(executors))
	if len(stepExecutors) > 0 {
		for _, se := range stepExecutors {
			rows = append(rows, models.ProcessStepExecutor{
				ProcessStepID:   stepID,
				EmployeeID:      se.EmployeeID,
				WorkloadPercent: se.WorkloadPercent,
			})
		}
	} else if len(executors) > 0 {
		for _, e := range executors {
			rows = append(rows, models.ProcessStepExecutor{
				ProcessStepID:   stepID,
				EmployeeID:      e.ID,
				WorkloadPercent: 0,
			})
		}
	}

	if len(rows) == 0 {
		return nil
	}

	return tx.Create(&rows).Error
}

func (r *ProcessRepository) GetAllFolders() ([]models.ProcessFolder, error) {

	var folders []models.ProcessFolder

	err := r.db.Find(&folders).Error

	return folders, err
}

func (r *ProcessRepository) GetAllProcesses() ([]models.Process, error) {

	var processes []models.Process

	err := r.db.Select("id", "name", "folder_id").Find(&processes).Error

	return processes, err
}

func (r *ProcessRepository) CreateFolder(folder *models.ProcessFolder) error {
	return r.db.Create(folder).Error
}

func (r *ProcessRepository) DeleteFolder(id uint) error {
	return r.db.Delete(&models.ProcessFolder{}, id).Error
}

func (r *ProcessRepository) UpdateFolder(folderID uint, name string, parentID *uint) error {
	if name == "" {
		return fmt.Errorf("folder name is required")
	}

	var folder models.ProcessFolder
	if err := r.db.Select("id").First(&folder, folderID).Error; err != nil {
		return err
	}

	if err := r.validateFolderMove(folderID, parentID); err != nil {
		return err
	}

	return r.db.Model(&models.ProcessFolder{}).
		Where("id = ?", folderID).
		Updates(map[string]interface{}{
			"name":      name,
			"parent_id": parentID,
		}).Error
}

func (r *ProcessRepository) MoveProcess(processID uint, folderID *uint) error {
	var process models.Process
	if err := r.db.Select("id").First(&process, processID).Error; err != nil {
		return err
	}

	if folderID != nil {
		var folder models.ProcessFolder
		if err := r.db.Select("id").First(&folder, *folderID).Error; err != nil {
			return err
		}
	}

	return r.db.Model(&models.Process{}).
		Where("id = ?", processID).
		Update("folder_id", folderID).Error
}

func (r *ProcessRepository) MoveFolder(folderID uint, parentID *uint) error {
	var folder models.ProcessFolder
	if err := r.db.Select("id").First(&folder, folderID).Error; err != nil {
		return err
	}

	if err := r.validateFolderMove(folderID, parentID); err != nil {
		return err
	}

	return r.db.Model(&models.ProcessFolder{}).
		Where("id = ?", folderID).
		Update("parent_id", parentID).Error
}

func (r *ProcessRepository) validateFolderMove(folderID uint, parentID *uint) error {
	if parentID != nil {
		if *parentID == folderID {
			return fmt.Errorf("folder cannot be parent of itself")
		}

		var parent models.ProcessFolder
		if err := r.db.Select("id").First(&parent, *parentID).Error; err != nil {
			return err
		}

		var all []models.ProcessFolder
		if err := r.db.Select("id", "parent_id").Find(&all).Error; err != nil {
			return err
		}

		parentMap := make(map[uint]*uint, len(all))
		for i := range all {
			parentMap[all[i].ID] = all[i].ParentID
		}

		cur := parentID
		for cur != nil {
			if *cur == folderID {
				return fmt.Errorf("folder move would create a cycle")
			}
			cur = parentMap[*cur]
		}
	}
	return nil
}
