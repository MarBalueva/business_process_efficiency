package controller

import (
	"errors"
	"net/http"
	"strconv"

	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StepController struct {
	service *service.ProcessService
}

func NewStepController(service *service.ProcessService) *StepController {
	return &StepController{service: service}
}

type ExecutorLoadRequest struct {
	EmployeeID      uint    `json:"employeeId" binding:"required"`
	WorkloadPercent float64 `json:"workloadPercent"`
}

type CreateStepRequest struct {
	ProcessVersionID uint                  `json:"processVersionId" binding:"required"`
	Name             string                `json:"name" binding:"required"`
	Type             string                `json:"type" binding:"required,oneof=START END INTERMEDIATE SUBPROCESS OPERATION CONDITION"`
	Description      string                `json:"description,omitempty"`
	ExecutorIDs      []uint                `json:"executorIds,omitempty"`
	ExecutorLoads    []ExecutorLoadRequest `json:"executorLoads,omitempty"`
}

type UpdateStepRequest struct {
	Name          string                `json:"Name" binding:"required"`
	Type          models.StepType       `json:"Type" binding:"required"`
	Description   string                `json:"Description,omitempty"`
	Executors     []models.Employee     `json:"Executors,omitempty"`
	ExecutorLoads []ExecutorLoadRequest `json:"ExecutorLoads,omitempty"`
	Metrics       *models.StepMetrics   `json:"Metrics,omitempty"`
}

// CreateStep godoc
// @Summary Создание шага процесса
// @Description Создает новый шаг для процесса
// @Tags process-steps
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param step body CreateStepRequest true "Step data"
// @Success 201 {object} models.ProcessStep
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /processes/steps [post]
func (h *StepController) CreateStep(c *gin.Context) {
	var req CreateStepRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var lastStep models.ProcessStep
	if err := h.service.GetLastStep(req.ProcessVersionID, &lastStep); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при получении последнего шага"})
		return
	}
	stepOrder := lastStep.StepOrder + 1

	step := models.ProcessStep{
		ProcessVersionID: req.ProcessVersionID,
		Name:             req.Name,
		Type:             models.StepType(req.Type),
		Description:      req.Description,
		StepOrder:        stepOrder,
	}

	if len(req.ExecutorLoads) > 0 {
		step.StepExecutors = make([]models.ProcessStepExecutor, 0, len(req.ExecutorLoads))
		for _, load := range req.ExecutorLoads {
			step.StepExecutors = append(step.StepExecutors, models.ProcessStepExecutor{
				EmployeeID:      load.EmployeeID,
				WorkloadPercent: load.WorkloadPercent,
			})
		}
	} else if len(req.ExecutorIDs) > 0 {
		step.Executors = make([]models.Employee, 0, len(req.ExecutorIDs))
		for _, employeeID := range req.ExecutorIDs {
			step.Executors = append(step.Executors, models.Employee{
				ID: employeeID,
			})
		}
	}

	if err := h.service.CreateStep(&step); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при создании шага"})
		return
	}

	c.JSON(http.StatusCreated, step)
}

// UpdateStep godoc
// @Summary Обновление шага процесса
// @Description Обновляет данные шага процесса по ID
// @Tags process-steps
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Step ID"
// @Param step body UpdateStepRequest true "Updated step data"
// @Success 200 {object} models.ProcessStep
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /processes/steps/{id} [put]
func (h *StepController) UpdateStep(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req UpdateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	step := models.ProcessStep{
		ID:          uint(id),
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Metrics:     req.Metrics,
	}

	if len(req.ExecutorLoads) > 0 {
		step.StepExecutors = make([]models.ProcessStepExecutor, 0, len(req.ExecutorLoads))
		for _, load := range req.ExecutorLoads {
			step.StepExecutors = append(step.StepExecutors, models.ProcessStepExecutor{
				EmployeeID:      load.EmployeeID,
				WorkloadPercent: load.WorkloadPercent,
			})
		}
	} else if len(req.Executors) > 0 {
		step.Executors = req.Executors
	}

	err := h.service.UpdateStep(&step)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.GetStep(step.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load updated step"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteStep godoc
// @Summary Удаление шага процесса
// @Description Удаляет шаг процесса по ID
// @Tags process-steps
// @Produce json
// @Security BearerAuth
// @Param id path int true "Step ID"
// @Success 200 {object} map[string]bool "deleted"
// @Failure 500 {object} map[string]string "error"
// @Router /processes/steps/{id} [delete]
func (h *StepController) DeleteStep(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	err := h.service.DeleteStep(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
