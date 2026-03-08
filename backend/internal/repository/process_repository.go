package repository

import (
	"business_process_efficiency/internal/models"

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
		Preload("Versions").
		Preload("Versions.Steps").
		Preload("Versions.Steps.Executors").
		First(&process, id).Error

	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (r *ProcessRepository) CreateProcess(process *models.Process) error {
	return r.db.Create(process).Error
}

func (r *ProcessRepository) UpdateProcess(process *models.Process) error {
	return r.db.Save(process).Error
}

func (r *ProcessRepository) DeleteProcess(id uint) error {
	return r.db.Delete(&models.Process{}, id).Error
}

func (r *ProcessRepository) CreateVersion(version *models.ProcessVersion) error {
	return r.db.Create(version).Error
}

func (r *ProcessRepository) DeleteVersion(id uint) error {
	return r.db.Delete(&models.ProcessVersion{}, id).Error
}

func (r *ProcessRepository) CreateStep(step *models.ProcessStep) error {
	return r.db.Create(step).Error
}

func (r *ProcessRepository) UpdateStep(step *models.ProcessStep) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(step).Error
}

func (r *ProcessRepository) DeleteStep(id uint) error {
	return r.db.Delete(&models.ProcessStep{}, id).Error
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
