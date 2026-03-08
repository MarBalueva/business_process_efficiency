package service

import (
	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/repository"
)

type ProcessService struct {
	repo *repository.ProcessRepository
}

func NewProcessService(repo *repository.ProcessRepository) *ProcessService {
	return &ProcessService{repo: repo}
}

func (s *ProcessService) GetRegistry() ([]models.ProcessFolder, error) {
	return s.repo.GetRegistry()
}

func (s *ProcessService) GetProcess(id uint) (*models.Process, error) {
	return s.repo.GetProcessByID(id)
}

func (s *ProcessService) CreateProcess(name string, folderID *uint, ownerID uint) (*models.Process, error) {

	process := models.Process{
		Name:     name,
		FolderID: folderID,
		OwnerID:  ownerID,
		IsActive: true,
	}

	err := s.repo.CreateProcess(&process)
	if err != nil {
		return nil, err
	}

	version := models.ProcessVersion{
		ProcessID: process.ID,
		Version:   1,
	}

	err = s.repo.CreateVersion(&version)
	if err != nil {
		return nil, err
	}

	return &process, nil
}

func (s *ProcessService) UpdateProcess(process *models.Process) error {
	return s.repo.UpdateProcess(process)
}

func (s *ProcessService) DeleteProcess(id uint) error {
	return s.repo.DeleteProcess(id)
}

func (s *ProcessService) CreateVersion(processID uint) (*models.ProcessVersion, error) {

	version := models.ProcessVersion{
		ProcessID: processID,
	}

	err := s.repo.CreateVersion(&version)

	return &version, err
}

func (s *ProcessService) DeleteVersion(id uint) error {
	return s.repo.DeleteVersion(id)
}

func (s *ProcessService) CreateStep(step *models.ProcessStep) error {
	return s.repo.CreateStep(step)
}

func (s *ProcessService) UpdateStep(step *models.ProcessStep) error {
	return s.repo.UpdateStep(step)
}

func (s *ProcessService) DeleteStep(id uint) error {
	return s.repo.DeleteStep(id)
}

func (s *ProcessService) GetRegistryTree() ([]*models.ProcessRegistryFolder, error) {

	folders, err := s.repo.GetAllFolders()
	if err != nil {
		return nil, err
	}

	processes, err := s.repo.GetAllProcesses()
	if err != nil {
		return nil, err
	}

	folderMap := map[uint]*models.ProcessRegistryFolder{}

	for _, f := range folders {

		folderMap[f.ID] = &models.ProcessRegistryFolder{
			ID:        f.ID,
			Name:      f.Name,
			ParentID:  f.ParentID,
			Processes: []models.ProcessShortDTO{},
			Children:  []*models.ProcessRegistryFolder{},
		}
	}

	// добавляем процессы
	for _, p := range processes {

		if p.FolderID == nil {
			continue
		}

		folder := folderMap[*p.FolderID]

		folder.Processes = append(folder.Processes, models.ProcessShortDTO{
			ID:   p.ID,
			Name: p.Name,
		})
	}

	var roots []*models.ProcessRegistryFolder

	for _, folder := range folderMap {

		if folder.ParentID == nil {
			roots = append(roots, folder)
			continue
		}

		parent := folderMap[*folder.ParentID]

		parent.Children = append(parent.Children, folder)
	}

	return roots, nil
}

func (s *ProcessService) CreateFolder(name string, parentID *uint) (*models.ProcessFolder, error) {
	folder := &models.ProcessFolder{
		Name:     name,
		ParentID: parentID,
	}

	err := s.repo.CreateFolder(folder)
	if err != nil {
		return nil, err
	}

	return folder, nil
}

func (s *ProcessService) DeleteFolder(id uint) error {
	return s.repo.DeleteFolder(id)
}
