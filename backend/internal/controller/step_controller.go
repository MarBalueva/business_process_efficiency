package controller

import (
	"net/http"
	"strconv"

	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/service"

	"github.com/gin-gonic/gin"
)

type StepController struct {
	service *service.ProcessService
}

func NewStepController(service *service.ProcessService) *StepController {
	return &StepController{service: service}
}

// CreateStep godoc
// @Summary Создание шага процесса
// @Description Создает новый шаг для процесса
// @Tags process-steps
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param step body models.ProcessStep true "Step data"
// @Success 201 {object} models.ProcessStep
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /steps [post]
func (h *StepController) CreateStep(c *gin.Context) {

	var step models.ProcessStep

	if err := c.ShouldBindJSON(&step); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	err := h.service.CreateStep(&step)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
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
// @Param step body models.ProcessStep true "Updated step data"
// @Success 200 {object} models.ProcessStep
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /steps/{id} [put]
func (h *StepController) UpdateStep(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))

	var step models.ProcessStep

	if err := c.ShouldBindJSON(&step); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	step.ID = uint(id)

	err := h.service.UpdateStep(&step)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, step)
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
// @Router /steps/{id} [delete]
func (h *StepController) DeleteStep(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))

	err := h.service.DeleteStep(uint(id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
