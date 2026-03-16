package service

import (
	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

type ProcessService struct {
	repo      *repository.ProcessRepository
	embedder  *EmbeddingClient
}

func NewProcessService(repo *repository.ProcessRepository, embedder *EmbeddingClient) *ProcessService {
	return &ProcessService{repo: repo, embedder: embedder}
}

func (s *ProcessService) GetRegistry() ([]models.ProcessFolder, error) {
	return s.repo.GetRegistry()
}

func (s *ProcessService) GetProcess(id uint) (*models.Process, error) {
	return s.repo.GetProcessByID(id)
}

func (s *ProcessService) GetStep(id uint) (*models.ProcessStep, error) {
	return s.repo.GetStepByID(id)
}

func (s *ProcessService) CreateProcess(name string, folderID *uint, ownerID uint, regularityCount int, regularityUnit string) (*models.Process, error) {
	if err := validateRegularity(regularityCount, regularityUnit); err != nil {
		return nil, err
	}

	process := models.Process{
		Name:            name,
		FolderID:        folderID,
		OwnerID:         ownerID,
		IsActive:        true,
		RegularityCount: regularityCount,
		RegularityUnit:  regularityUnit,
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

func (s *ProcessService) UpdateProcess(id uint, req models.UpdateProcessRequest) (*models.Process, error) {
	if err := validateRegularity(req.RegularityCount, req.RegularityUnit); err != nil {
		return nil, err
	}

	process, err := s.repo.UpdateProcess(id, req)
	if err != nil {
		return nil, err
	}

	return process, nil
}

func (s *ProcessService) DeleteProcess(id uint) error {
	return s.repo.DeleteProcess(id)
}

func (s *ProcessService) CreateVersion(processID uint) (*models.ProcessVersion, error) {

	lastVersion, err := s.repo.GetLastVersionNumber(processID)
	if err != nil {
		return nil, err
	}

	version := models.ProcessVersion{
		ProcessID: processID,
		Version:   lastVersion + 1,
	}

	err = s.repo.CreateVersion(&version)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

func (s *ProcessService) DeleteVersion(id uint) error {
	return s.repo.DeleteVersion(id)
}

func (s *ProcessService) CreateStep(step *models.ProcessStep) error {
	if err := validateStepRules(step); err != nil {
		return err
	}
	if err := s.repo.CreateStep(step); err != nil {
		return err
	}
	_ = s.reindexStepSemantic(context.Background(), step.ID)
	return nil
}

func (s *ProcessService) GetLastStep(versionID uint, step *models.ProcessStep) error {
	return s.repo.GetLastStepByVersion(versionID, step)
}

func (s *ProcessService) GetEmployeesByIDs(ids []uint, employees *[]models.Employee) error {
	return s.repo.GetEmployeesByIDs(ids, employees)
}

func (s *ProcessService) UpdateStep(step *models.ProcessStep) error {
	if err := validateStepRules(step); err != nil {
		return err
	}
	if err := s.repo.UpdateStep(step); err != nil {
		return err
	}
	_ = s.reindexStepSemantic(context.Background(), step.ID)
	return nil
}

func (s *ProcessService) DeleteStep(id uint) error {
	if err := s.repo.DeleteStep(id); err != nil {
		return err
	}
	_ = s.repo.DeleteStepSemanticIndex(id)
	return nil
}

func (s *ProcessService) ReorderSteps(processVersionID uint, orderedStepIDs []uint) error {
	return s.repo.ReorderSteps(processVersionID, orderedStepIDs)
}

func (s *ProcessService) SuggestSteps(ctx context.Context, query string, excludeProcessID *uint, limit int) ([]models.StepSuggestion, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []models.StepSuggestion{}, nil
	}
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	candidates, err := s.repo.ListStepSemanticCandidates(query, excludeProcessID, limit)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return []models.StepSuggestion{}, nil
	}

	queryVec, err := s.embedOne(ctx, query)
	if err != nil {
		// fallback: lexical ranking only
		out := make([]models.StepSuggestion, 0, min(limit, len(candidates)))
		for i, c := range candidates {
			if i >= limit {
				break
			}
			out = append(out, models.StepSuggestion{
				StepID:    c.StepID,
				ProcessID: c.ProcessID,
				StepName:  c.StepName,
				StepType:  c.StepType,
				Score:     0,
			})
		}
		return out, nil
	}

	type scored struct {
		row   models.StepSemanticIndex
		score float64
	}
	scoredRows := make([]scored, 0, len(candidates))
	for _, c := range candidates {
		vec, err := parseEmbeddingJSON(c.EmbeddingJSON)
		if err != nil || len(vec) == 0 {
			continue
		}
		score := cosineSimilarity(queryVec, vec)
		scoredRows = append(scoredRows, scored{row: c, score: score})
	}
	if len(scoredRows) == 0 {
		return []models.StepSuggestion{}, nil
	}

	sort.Slice(scoredRows, func(i, j int) bool {
		return scoredRows[i].score > scoredRows[j].score
	})

	out := make([]models.StepSuggestion, 0, min(limit, len(scoredRows)))
	for i, item := range scoredRows {
		if i >= limit {
			break
		}
		out = append(out, models.StepSuggestion{
			StepID:    item.row.StepID,
			ProcessID: item.row.ProcessID,
			StepName:  item.row.StepName,
			StepType:  item.row.StepType,
			Score:     item.score,
		})
	}
	return out, nil
}

func (s *ProcessService) ReindexAllSteps(ctx context.Context) (int, int, string, error) {
	ids, err := s.repo.ListAllStepIDs()
	if err != nil {
		return 0, 0, "", err
	}
	count := 0
	failed := 0
	lastErr := ""
	for _, id := range ids {
		if err := s.reindexStepSemantic(ctx, id); err != nil {
			failed++
			lastErr = err.Error()
			continue
		}
		count++
	}
	return count, failed, lastErr, nil
}

func (s *ProcessService) reindexStepSemantic(ctx context.Context, stepID uint) error {
	source, err := s.repo.GetStepIndexSource(stepID)
	if err != nil {
		return err
	}
	vector, err := s.embedOne(ctx, source.StepName)
	if err != nil {
		return err
	}
	return s.repo.UpsertStepSemanticIndex(source.StepID, source.ProcessID, source.StepType, source.StepName, vector)
}

func (s *ProcessService) embedOne(ctx context.Context, text string) ([]float64, error) {
	if s.embedder == nil || !s.embedder.Enabled() {
		return nil, fmt.Errorf("embedding disabled")
	}
	vectors, err := s.embedder.EmbedTexts(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) != 1 {
		return nil, fmt.Errorf("invalid embedding response")
	}
	return vectors[0], nil
}

func parseEmbeddingJSON(raw string) ([]float64, error) {
	var vec []float64
	if err := json.Unmarshal([]byte(raw), &vec); err != nil {
		return nil, err
	}
	return vec, nil
}

func cosineSimilarity(a []float64, b []float64) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}

	dot := 0.0
	normA := 0.0
	normB := 0.0
	for i := 0; i < n; i++ {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
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

func (s *ProcessService) UpdateFolder(folderID uint, name string, parentID *uint) error {
	return s.repo.UpdateFolder(folderID, name, parentID)
}

func (s *ProcessService) MoveProcess(processID uint, folderID *uint) error {
	return s.repo.MoveProcess(processID, folderID)
}

func (s *ProcessService) MoveFolder(folderID uint, parentID *uint) error {
	return s.repo.MoveFolder(folderID, parentID)
}

func validateStepRules(step *models.ProcessStep) error {
	isExecutorAllowed := step.Type == models.StepOperation || step.Type == models.StepSubprocess
	if !isExecutorAllowed && (len(step.StepExecutors) > 0 || len(step.Executors) > 0) {
		return fmt.Errorf("executors are allowed only for OPERATION and SUBPROCESS")
	}

	if len(step.ParallelSteps) > 0 {
		return fmt.Errorf("parallel steps are deprecated; use PARALLEL_GATEWAY with parallel branches")
	}

	if step.Type == models.StepParallelGateway {
		if len(step.ParallelBranches) > 0 {
			seenParallelBranches := make(map[uint]struct{}, len(step.ParallelBranches))
			for _, b := range step.ParallelBranches {
				if b.NextStepID == 0 {
					return fmt.Errorf("parallel branch nextStepId is required")
				}
				if _, ok := seenParallelBranches[b.NextStepID]; ok {
					return fmt.Errorf("parallel branch nextStepId %d duplicated", b.NextStepID)
				}
				seenParallelBranches[b.NextStepID] = struct{}{}
			}
		}
	} else if len(step.ParallelBranches) > 0 {
		return fmt.Errorf("parallel branches are allowed only for PARALLEL_GATEWAY")
	}

	if step.Type == models.StepCondition {
		if len(step.ConditionBranches) == 0 {
			return nil
		}

		seen := make(map[uint]struct{}, len(step.ConditionBranches))
		sum := 0.0
		for _, b := range step.ConditionBranches {
			if b.NextStepID == 0 {
				return fmt.Errorf("condition branch nextStepId is required")
			}
			if _, ok := seen[b.NextStepID]; ok {
				return fmt.Errorf("condition branch nextStepId %d duplicated", b.NextStepID)
			}
			seen[b.NextStepID] = struct{}{}
			if b.ProbabilityPercent < 0 || b.ProbabilityPercent > 100 {
				return fmt.Errorf("condition branch probability for step %d must be in range 0..100", b.NextStepID)
			}
			sum += b.ProbabilityPercent
		}
		if math.Abs(sum-100) > 0.0001 {
			return fmt.Errorf("sum of condition branch probabilities must be exactly 100")
		}
	} else if len(step.ConditionBranches) > 0 {
		return fmt.Errorf("condition branches are allowed only for CONDITION")
	}

	if step.Type == models.StepParallelEnd {
		if step.ClosesStepID == nil || *step.ClosesStepID == 0 {
			return fmt.Errorf("parallel end must reference closesStepId")
		}
		if len(step.PreviousSteps) == 0 {
			return fmt.Errorf("parallel end must contain previous steps")
		}
	}
	if step.Type == models.StepConditionEnd {
		if step.ClosesStepID == nil || *step.ClosesStepID == 0 {
			return fmt.Errorf("condition end must reference closesStepId")
		}
		if len(step.PreviousSteps) == 0 {
			return fmt.Errorf("condition end must contain previous steps")
		}
	}

	if len(step.StepExecutors) == 0 {
		return nil
	}

	sum := 0.0
	seen := make(map[uint]struct{}, len(step.StepExecutors))

	for _, se := range step.StepExecutors {
		if se.EmployeeID == 0 {
			return fmt.Errorf("executor employeeId is required")
		}
		if _, exists := seen[se.EmployeeID]; exists {
			return fmt.Errorf("executor %d duplicated", se.EmployeeID)
		}
		seen[se.EmployeeID] = struct{}{}

		if se.WorkloadPercent < 0 || se.WorkloadPercent > 100 {
			return fmt.Errorf("workload percent for executor %d must be in range 0..100", se.EmployeeID)
		}

		sum += se.WorkloadPercent
	}

	if math.Abs(sum-100) > 0.0001 {
		return fmt.Errorf("sum of executor workload percents must be exactly 100")
	}

	return nil
}

func validateRegularity(count int, unit string) error {
	if count < 0 {
		return errors.New("regularity_count must be >= 0")
	}

	normalized := strings.TrimSpace(strings.ToLower(unit))
	if count == 0 {
		if normalized != "" {
			return errors.New("regularity_unit must be empty when regularity_count is 0")
		}
		return nil
	}

	switch normalized {
	case "day", "week", "month", "quarter", "halfyear", "year":
		return nil
	default:
		return errors.New("regularity_unit must be one of: day, week, month, quarter, halfyear, year")
	}
}
