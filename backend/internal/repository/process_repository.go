package repository

import (
	"business_process_efficiency/internal/models"
	"encoding/json"
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

type StepIndexSource struct {
	StepID      uint
	ProcessID   uint
	ProcessName string
	StepName    string
	StepType    models.StepType
}

type StepSuggestionCandidate struct {
	StepID      uint
	ProcessID   uint
	ProcessName string
	StepName    string
	StepType    models.StepType
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
		Preload("Versions.Steps.ParallelSteps").
		Preload("Versions.Steps.ParallelBranches").
		Preload("Versions.Steps.ConditionBranches").
		Preload("Versions.Steps.PreviousSteps").
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
		Preload("ParallelSteps").
		Preload("ParallelBranches").
		Preload("ConditionBranches").
		Preload("PreviousSteps").
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
		ClosesStepID:     step.ClosesStepID,
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
	if err := r.replaceStepParallelsTx(tx, step.ID, step.ParallelSteps); err != nil {
		tx.Rollback()
		return err
	}
	if err := r.replaceParallelBranchesTx(tx, step.ID, step.ParallelBranches, step.Type); err != nil {
		tx.Rollback()
		return err
	}
	if err := r.replaceConditionBranchesTx(tx, step.ID, step.ConditionBranches, step.Type); err != nil {
		tx.Rollback()
		return err
	}
	if err := r.replaceStepPreviousTx(tx, step.ID, step.PreviousSteps); err != nil {
		tx.Rollback()
		return err
	}
	if err := r.syncAutoClosingStepTx(tx, base); err != nil {
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
	if step.ClosesStepID != nil || step.Type == models.StepParallelEnd || step.Type == models.StepConditionEnd {
		updateFields["closes_step_id"] = step.ClosesStepID
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
	if step.ParallelSteps != nil {
		if err := r.replaceStepParallelsTx(tx, step.ID, step.ParallelSteps); err != nil {
			tx.Rollback()
			return err
		}
	}
	if step.ParallelBranches != nil {
		if err := r.replaceParallelBranchesTx(tx, step.ID, step.ParallelBranches, step.Type); err != nil {
			tx.Rollback()
			return err
		}
	} else if step.Type != models.StepParallelGateway {
		if err := tx.Where("gateway_step_id = ?", step.ID).Delete(&models.ProcessParallelBranch{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if step.ConditionBranches != nil {
		if err := r.replaceConditionBranchesTx(tx, step.ID, step.ConditionBranches, step.Type); err != nil {
			tx.Rollback()
			return err
		}
	} else if step.Type != models.StepCondition {
		if err := tx.Where("condition_step_id = ?", step.ID).Delete(&models.ProcessConditionBranch{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if step.PreviousSteps != nil {
		if err := r.replaceStepPreviousTx(tx, step.ID, step.PreviousSteps); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := r.syncAutoClosingStepTx(tx, models.ProcessStep{
		ID:               step.ID,
		ProcessVersionID: currentStep.ProcessVersionID,
		Name:             step.Name,
		Type:             step.Type,
	}); err != nil {
		tx.Rollback()
		return err
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
	if err := tx.Where("process_step_id = ? OR parallel_step_id = ?", id, id).Delete(&models.ProcessStepParallel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("condition_step_id = ? OR next_step_id = ?", id, id).Delete(&models.ProcessConditionBranch{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("gateway_step_id = ? OR next_step_id = ?", id, id).Delete(&models.ProcessParallelBranch{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("step_id = ? OR previous_step_id = ?", id, id).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	var autoClosing []models.ProcessStep
	if err := tx.Select("id").
		Where("closes_step_id = ? AND (type = ? OR type = ?)", id, models.StepParallelEnd, models.StepConditionEnd).
		Find(&autoClosing).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, c := range autoClosing {
		if err := tx.Where("step_id = ? OR previous_step_id = ?", c.ID, c.ID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Delete(&models.ProcessStep{}, c.ID).Error; err != nil {
			tx.Rollback()
			return err
		}
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

func (r *ProcessRepository) ReorderSteps(processVersionID uint, orderedStepIDs []uint) error {
	if processVersionID == 0 {
		return errors.New("processVersionId is required")
	}
	if len(orderedStepIDs) == 0 {
		return errors.New("orderedStepIds must not be empty")
	}

	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var existingSteps []models.ProcessStep
	if err := tx.
		Select("id", "step_order").
		Where("process_version_id = ?", processVersionID).
		Order("step_order ASC, id ASC").
		Find(&existingSteps).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(existingSteps) != len(orderedStepIDs) {
		tx.Rollback()
		return fmt.Errorf("orderedStepIds count mismatch: expected %d, got %d", len(existingSteps), len(orderedStepIDs))
	}

	existingSet := make(map[uint]struct{}, len(existingSteps))
	for _, s := range existingSteps {
		existingSet[s.ID] = struct{}{}
	}

	seen := make(map[uint]struct{}, len(orderedStepIDs))
	for _, id := range orderedStepIDs {
		if _, ok := existingSet[id]; !ok {
			tx.Rollback()
			return fmt.Errorf("step %d does not belong to process version %d", id, processVersionID)
		}
		if _, duplicated := seen[id]; duplicated {
			tx.Rollback()
			return fmt.Errorf("step %d duplicated in orderedStepIds", id)
		}
		seen[id] = struct{}{}
	}

	for idx, id := range orderedStepIDs {
		nextOrder := idx + 1
		if err := tx.Model(&models.ProcessStep{}).
			Where("id = ? AND process_version_id = ?", id, processVersionID).
			Update("step_order", nextOrder).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := r.rebuildStepLinksTx(tx, processVersionID); err != nil {
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

func (r *ProcessRepository) replaceStepParallelsTx(
	tx *gorm.DB,
	stepID uint,
	parallels []models.ProcessStepParallel,
) error {
	var oldRows []models.ProcessStepParallel
	if err := tx.Where("process_step_id = ?", stepID).Find(&oldRows).Error; err != nil {
		return err
	}
	oldSet := make(map[uint]struct{}, len(oldRows))
	for _, row := range oldRows {
		oldSet[row.ParallelStepID] = struct{}{}
	}

	if err := tx.Where("process_step_id = ?", stepID).Delete(&models.ProcessStepParallel{}).Error; err != nil {
		return err
	}

	newSet := make(map[uint]struct{}, len(parallels))
	for _, p := range parallels {
		if p.ParallelStepID == 0 || p.ParallelStepID == stepID {
			continue
		}
		newSet[p.ParallelStepID] = struct{}{}
	}

	for parallelID := range newSet {
		row := models.ProcessStepParallel{ProcessStepID: stepID, ParallelStepID: parallelID}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}

		var reverse models.ProcessStepParallel
		err := tx.Where("process_step_id = ? AND parallel_step_id = ?", parallelID, stepID).First(&reverse).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&models.ProcessStepParallel{ProcessStepID: parallelID, ParallelStepID: stepID}).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	for oldParallelID := range oldSet {
		if _, still := newSet[oldParallelID]; still {
			continue
		}
		if err := tx.Where("process_step_id = ? AND parallel_step_id = ?", oldParallelID, stepID).Delete(&models.ProcessStepParallel{}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *ProcessRepository) replaceConditionBranchesTx(
	tx *gorm.DB,
	stepID uint,
	branches []models.ProcessConditionBranch,
	stepType models.StepType,
) error {
	var oldRows []models.ProcessConditionBranch
	if err := tx.Where("condition_step_id = ?", stepID).Find(&oldRows).Error; err != nil {
		return err
	}
	oldNext := make(map[uint]struct{}, len(oldRows))
	for _, row := range oldRows {
		oldNext[row.NextStepID] = struct{}{}
	}

	if err := tx.Where("condition_step_id = ?", stepID).Delete(&models.ProcessConditionBranch{}).Error; err != nil {
		return err
	}

	if stepType != models.StepCondition || len(branches) == 0 {
		for nextID := range oldNext {
			if err := tx.Where("step_id = ? AND previous_step_id = ?", nextID, stepID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
				return err
			}
		}
		return nil
	}

	rows := make([]models.ProcessConditionBranch, 0, len(branches))
	newNext := make(map[uint]struct{}, len(branches))
	for _, b := range branches {
		if b.NextStepID == 0 {
			continue
		}
		newNext[b.NextStepID] = struct{}{}
		rows = append(rows, models.ProcessConditionBranch{
			ConditionStepID:    stepID,
			NextStepID:         b.NextStepID,
			ProbabilityPercent: b.ProbabilityPercent,
		})
	}
	if len(rows) == 0 {
		for nextID := range oldNext {
			if err := tx.Where("step_id = ? AND previous_step_id = ?", nextID, stepID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
				return err
			}
		}
		return nil
	}
	if err := tx.Create(&rows).Error; err != nil {
		return err
	}

	for nextID := range oldNext {
		if _, still := newNext[nextID]; still {
			continue
		}
		if err := tx.Where("step_id = ? AND previous_step_id = ?", nextID, stepID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
			return err
		}
	}
	for nextID := range newNext {
		if err := tx.Where("step_id = ?", nextID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.ProcessStepPrevious{StepID: nextID, PreviousStepID: stepID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *ProcessRepository) replaceParallelBranchesTx(
	tx *gorm.DB,
	stepID uint,
	branches []models.ProcessParallelBranch,
	stepType models.StepType,
) error {
	var oldRows []models.ProcessParallelBranch
	if err := tx.Where("gateway_step_id = ?", stepID).Find(&oldRows).Error; err != nil {
		return err
	}
	oldNext := make(map[uint]struct{}, len(oldRows))
	for _, row := range oldRows {
		oldNext[row.NextStepID] = struct{}{}
	}

	if err := tx.Where("gateway_step_id = ?", stepID).Delete(&models.ProcessParallelBranch{}).Error; err != nil {
		return err
	}

	if stepType != models.StepParallelGateway || len(branches) == 0 {
		for nextID := range oldNext {
			if err := tx.Where("step_id = ? AND previous_step_id = ?", nextID, stepID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
				return err
			}
		}
		return nil
	}

	rows := make([]models.ProcessParallelBranch, 0, len(branches))
	newNext := make(map[uint]struct{}, len(branches))
	for _, b := range branches {
		if b.NextStepID == 0 {
			continue
		}
		newNext[b.NextStepID] = struct{}{}
		rows = append(rows, models.ProcessParallelBranch{
			GatewayStepID: stepID,
			NextStepID:    b.NextStepID,
		})
	}
	if len(rows) == 0 {
		for nextID := range oldNext {
			if err := tx.Where("step_id = ? AND previous_step_id = ?", nextID, stepID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
				return err
			}
		}
		return nil
	}
	if err := tx.Create(&rows).Error; err != nil {
		return err
	}

	for nextID := range oldNext {
		if _, still := newNext[nextID]; still {
			continue
		}
		if err := tx.Where("step_id = ? AND previous_step_id = ?", nextID, stepID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
			return err
		}
	}
	for nextID := range newNext {
		if err := tx.Where("step_id = ?", nextID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.ProcessStepPrevious{StepID: nextID, PreviousStepID: stepID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *ProcessRepository) replaceStepPreviousTx(
	tx *gorm.DB,
	stepID uint,
	previous []models.ProcessStepPrevious,
) error {
	var oldRows []models.ProcessStepPrevious
	if err := tx.Where("step_id = ?", stepID).Find(&oldRows).Error; err != nil {
		return err
	}
	oldPrevIDs := make([]uint, 0, len(oldRows))
	for _, row := range oldRows {
		oldPrevIDs = append(oldPrevIDs, row.PreviousStepID)
	}

	if err := tx.Where("step_id = ?", stepID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
		return err
	}
	if err := r.removeImplicitBranchesByPreviousTx(tx, stepID, oldPrevIDs); err != nil {
		return err
	}

	if len(previous) == 0 {
		return nil
	}
	rows := make([]models.ProcessStepPrevious, 0, len(previous))
	for _, p := range previous {
		if p.PreviousStepID == 0 {
			continue
		}
		rows = append(rows, models.ProcessStepPrevious{
			StepID:         stepID,
			PreviousStepID: p.PreviousStepID,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	if err := tx.Create(&rows).Error; err != nil {
		return err
	}
	newPrevIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		newPrevIDs = append(newPrevIDs, row.PreviousStepID)
	}
	if err := r.ensureImplicitBranchesByPreviousTx(tx, stepID, newPrevIDs); err != nil {
		return err
	}

	// If user inserted a regular step between two linear steps, move follower:
	// old follower(previous -> X) becomes follower(newStep -> X).
	if len(rows) == 1 {
		prevID := rows[0].PreviousStepID
		if err := r.rewireLinearFollowerTx(tx, stepID, prevID); err != nil {
			return err
		}
	}

	return nil
}

func (r *ProcessRepository) removeImplicitBranchesByPreviousTx(tx *gorm.DB, stepID uint, prevIDs []uint) error {
	seen := make(map[uint]struct{}, len(prevIDs))
	for _, prevID := range prevIDs {
		if prevID == 0 {
			continue
		}
		if _, ok := seen[prevID]; ok {
			continue
		}
		seen[prevID] = struct{}{}

		var prevStep models.ProcessStep
		if err := tx.Select("id", "type").First(&prevStep, prevID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}

		switch prevStep.Type {
		case models.StepCondition:
			if err := tx.Where("condition_step_id = ? AND next_step_id = ?", prevID, stepID).
				Delete(&models.ProcessConditionBranch{}).Error; err != nil {
				return err
			}
		case models.StepParallelGateway:
			if err := tx.Where("gateway_step_id = ? AND next_step_id = ?", prevID, stepID).
				Delete(&models.ProcessParallelBranch{}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ProcessRepository) ensureImplicitBranchesByPreviousTx(tx *gorm.DB, stepID uint, prevIDs []uint) error {
	seen := make(map[uint]struct{}, len(prevIDs))
	for _, prevID := range prevIDs {
		if prevID == 0 {
			continue
		}
		if _, ok := seen[prevID]; ok {
			continue
		}
		seen[prevID] = struct{}{}

		var prevStep models.ProcessStep
		if err := tx.Select("id", "type").First(&prevStep, prevID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}

		switch prevStep.Type {
		case models.StepCondition:
			var count int64
			if err := tx.Model(&models.ProcessConditionBranch{}).
				Where("condition_step_id = ? AND next_step_id = ?", prevID, stepID).
				Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				row := models.ProcessConditionBranch{
					ConditionStepID:   prevID,
					NextStepID:        stepID,
					ProbabilityPercent: 0,
				}
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
			}
		case models.StepParallelGateway:
			var count int64
			if err := tx.Model(&models.ProcessParallelBranch{}).
				Where("gateway_step_id = ? AND next_step_id = ?", prevID, stepID).
				Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				row := models.ProcessParallelBranch{
					GatewayStepID: prevID,
					NextStepID:    stepID,
				}
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *ProcessRepository) rewireLinearFollowerTx(tx *gorm.DB, newStepID uint, prevID uint) error {
	var prevStep models.ProcessStep
	if err := tx.Select("id", "process_version_id", "type").First(&prevStep, prevID).Error; err != nil {
		return err
	}

	// Branching sources are handled by explicit branch relations; do not auto-rewire them.
	if prevStep.Type == models.StepCondition || prevStep.Type == models.StepParallelGateway {
		return nil
	}

	type followerRow struct {
		StepID uint
		Type   models.StepType
	}
	var followers []followerRow
	if err := tx.Table("process_step_previous p").
		Select("p.step_id as step_id, s.type as type").
		Joins("JOIN process_steps s ON s.id = p.step_id").
		Where("p.previous_step_id = ? AND p.step_id <> ? AND s.process_version_id = ?", prevID, newStepID, prevStep.ProcessVersionID).
		Order("s.step_order ASC, s.id ASC").
		Scan(&followers).Error; err != nil {
		return err
	}
	if len(followers) == 0 {
		return nil
	}

	f := followers[0]
	// Closing steps can have multiple previous links.
	// Replace only this particular branch edge: prevID -> closing to newStepID -> closing.
	if f.Type == models.StepConditionEnd || f.Type == models.StepParallelEnd {
		return tx.Model(&models.ProcessStepPrevious{}).
			Where("step_id = ? AND previous_step_id = ?", f.StepID, prevID).
			Update("previous_step_id", newStepID).Error
	}

	if err := tx.Where("step_id = ?", f.StepID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
		return err
	}

	return tx.Create(&models.ProcessStepPrevious{
		StepID:         f.StepID,
		PreviousStepID: newStepID,
	}).Error
}

func (r *ProcessRepository) syncAutoClosingStepTx(tx *gorm.DB, source models.ProcessStep) error {
	switch source.Type {
	case models.StepParallelGateway:
		var branches []models.ProcessParallelBranch
		if err := tx.Where("gateway_step_id = ?", source.ID).Find(&branches).Error; err != nil {
			return err
		}
		if len(branches) == 0 {
			return r.deleteClosingStepsBySourceTx(tx, source.ID, models.StepParallelEnd)
		}
		previous := make([]models.ProcessStepPrevious, 0, len(branches))
		for _, b := range branches {
			previous = append(previous, models.ProcessStepPrevious{PreviousStepID: b.NextStepID})
		}
		return r.upsertClosingStepTx(tx, source, models.StepParallelEnd, "Конец параллели", previous)
	case models.StepCondition:
		var branches []models.ProcessConditionBranch
		if err := tx.Where("condition_step_id = ?", source.ID).Find(&branches).Error; err != nil {
			return err
		}
		if len(branches) == 0 {
			return r.deleteClosingStepsBySourceTx(tx, source.ID, models.StepConditionEnd)
		}
		previous := make([]models.ProcessStepPrevious, 0, len(branches))
		for _, b := range branches {
			previous = append(previous, models.ProcessStepPrevious{PreviousStepID: b.NextStepID})
		}
		return r.upsertClosingStepTx(tx, source, models.StepConditionEnd, "Конец условия", previous)
	default:
		var closing []models.ProcessStep
		if err := tx.Select("id", "type").
			Where("closes_step_id = ? AND (type = ? OR type = ?)", source.ID, models.StepParallelEnd, models.StepConditionEnd).
			Find(&closing).Error; err != nil {
			return err
		}
		for _, c := range closing {
			if err := tx.Where("step_id = ? OR previous_step_id = ?", c.ID, c.ID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.ProcessStep{}, c.ID).Error; err != nil {
				return err
			}
		}
		return nil
	}
}

func (r *ProcessRepository) deleteClosingStepsBySourceTx(tx *gorm.DB, sourceID uint, closingType models.StepType) error {
	var closing []models.ProcessStep
	if err := tx.Select("id").
		Where("closes_step_id = ? AND type = ?", sourceID, closingType).
		Find(&closing).Error; err != nil {
		return err
	}
	for _, c := range closing {
		if err := tx.Where("step_id = ? OR previous_step_id = ?", c.ID, c.ID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.ProcessStep{}, c.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *ProcessRepository) upsertClosingStepTx(
	tx *gorm.DB,
	source models.ProcessStep,
	closingType models.StepType,
	closingNamePrefix string,
	previous []models.ProcessStepPrevious,
) error {
	var closings []models.ProcessStep
	if err := tx.Where("closes_step_id = ? AND type = ?", source.ID, closingType).Order("id ASC").Find(&closings).Error; err != nil {
		return err
	}

	var closing models.ProcessStep
	hasClosing := len(closings) > 0
	if hasClosing {
		closing = closings[0]
		for i := 1; i < len(closings); i++ {
			extraID := closings[i].ID
			if err := tx.Where("step_id = ? OR previous_step_id = ?", extraID, extraID).Delete(&models.ProcessStepPrevious{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.ProcessStep{}, extraID).Error; err != nil {
				return err
			}
		}
	}

	if !hasClosing {
		var last models.ProcessStep
		if err := tx.Where("process_version_id = ?", source.ProcessVersionID).Order("step_order DESC, id DESC").First(&last).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		closing = models.ProcessStep{
			ProcessVersionID: source.ProcessVersionID,
			StepOrder:        last.StepOrder + 1,
			Name:             fmt.Sprintf("%s: %s", closingNamePrefix, source.Name),
			Type:             closingType,
			Description:      "",
			ClosesStepID:     &source.ID,
		}
		if err := tx.Create(&closing).Error; err != nil {
			return err
		}
	} else {
		if err := tx.Model(&models.ProcessStep{}).Where("id = ?", closing.ID).Updates(map[string]interface{}{
			"name": fmt.Sprintf("%s: %s", closingNamePrefix, source.Name),
		}).Error; err != nil {
			return err
		}
	}

	return r.replaceStepPreviousTx(tx, closing.ID, previous)
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

func (r *ProcessRepository) GetStepIndexSource(stepID uint) (*StepIndexSource, error) {
	var row StepIndexSource
	err := r.db.Table("process_steps s").
		Select("s.id as step_id, s.name as step_name, s.type as step_type, pv.process_id as process_id, p.name as process_name").
		Joins("JOIN process_versions pv ON pv.id = s.process_version_id").
		Joins("JOIN processes p ON p.id = pv.process_id").
		Where("s.id = ?", stepID).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.StepID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *ProcessRepository) UpsertStepSemanticIndex(stepID uint, processID uint, stepType models.StepType, stepName string, vector []float64) error {
	raw, err := json.Marshal(vector)
	if err != nil {
		return err
	}
	row := models.StepSemanticIndex{
		StepID:        stepID,
		ProcessID:     processID,
		StepType:      stepType,
		StepName:      stepName,
		EmbeddingJSON: string(raw),
	}
	return r.db.
		Where("step_id = ?", stepID).
		Assign(row).
		FirstOrCreate(&models.StepSemanticIndex{}).Error
}

func (r *ProcessRepository) DeleteStepSemanticIndex(stepID uint) error {
	return r.db.Where("step_id = ?", stepID).Delete(&models.StepSemanticIndex{}).Error
}

func (r *ProcessRepository) ListStepSemanticCandidates(_ string, excludeProcessID *uint, limit int) ([]models.StepSemanticIndex, error) {
	if limit <= 0 {
		limit = 5
	}
	candidateLimit := limit * 20
	if candidateLimit < 100 {
		candidateLimit = 100
	}
	if candidateLimit > 400 {
		candidateLimit = 400
	}

	q := r.db.Model(&models.StepSemanticIndex{})
	if excludeProcessID != nil && *excludeProcessID > 0 {
		q = q.Where("process_id <> ?", *excludeProcessID)
	}

	var rows []models.StepSemanticIndex
	err := q.Order("updated_at DESC").Limit(candidateLimit).Find(&rows).Error
	return rows, err
}

func (r *ProcessRepository) ListAllStepIDs() ([]uint, error) {
	var ids []uint
	err := r.db.Model(&models.ProcessStep{}).Pluck("id", &ids).Error
	return ids, err
}

func (r *ProcessRepository) ListStepSuggestionCandidates(query string, excludeProcessID *uint, limit int) ([]StepSuggestionCandidate, error) {
	if limit <= 0 {
		limit = 5
	}
	candidateLimit := limit * 20
	if candidateLimit < 60 {
		candidateLimit = 60
	}
	if candidateLimit > 300 {
		candidateLimit = 300
	}

	q := r.db.Table("process_steps s").
		Select("s.id as step_id, pv.process_id as process_id, p.name as process_name, s.name as step_name, s.type as step_type").
		Joins("JOIN process_versions pv ON pv.id = s.process_version_id").
		Joins("JOIN processes p ON p.id = pv.process_id")

	if excludeProcessID != nil && *excludeProcessID > 0 {
		q = q.Where("pv.process_id <> ?", *excludeProcessID)
	}
	if query != "" {
		q = q.Where("s.name ILIKE ?", "%"+query+"%")
	}

	var rows []StepSuggestionCandidate
	err := q.Order("s.updated_at DESC, s.id DESC").Limit(candidateLimit).Scan(&rows).Error
	return rows, err
}
