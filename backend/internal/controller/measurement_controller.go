package controller

import (
	"net/http"
	"strconv"

	"business_process_efficiency/internal/service"

	"github.com/gin-gonic/gin"
)

type MeasurementController struct {
	service *service.MeasurementService
}

func NewMeasurementController(service *service.MeasurementService) *MeasurementController {
	return &MeasurementController{service: service}
}

// StartMeasurement godoc
// @Summary Начало измерения
// @Description Начинает измерение времени для шага
// @Tags measurements
// @Produce json
// @Security BearerAuth
// @Param stepId query int true "Step ID"
// @Success 200 {object} models.StepMeasurement
// @Failure 500 {object} map[string]string "error"
// @Router /measurements/start [post]
func (h *MeasurementController) Start(c *gin.Context) {

	stepID, _ := strconv.Atoi(c.Query("stepId"))

	m, err := h.service.StartMeasurement(uint(stepID))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, m)
}

// ListMeasurements godoc
// @Summary Список замеров шага
// @Description Возвращает замеры времени для шага по stepId
// @Tags measurements
// @Produce json
// @Security BearerAuth
// @Param stepId query int true "Step ID"
// @Success 200 {array} models.StepMeasurement
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /measurements [get]
func (h *MeasurementController) List(c *gin.Context) {
	stepID, err := strconv.Atoi(c.Query("stepId"))
	if err != nil || stepID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stepId"})
		return
	}

	list, err := h.service.GetMeasurementsByStep(uint(stepID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, list)
}

// PauseMeasurement godoc
// @Summary Пауза измерения
// @Description Ставит измерение на паузу
// @Tags measurements
// @Produce json
// @Security BearerAuth
// @Param measurementId query int true "Measurement ID"
// @Success 200 {object} map[string]bool "paused"
// @Failure 500 {object} map[string]string "error"
// @Router /measurements/pause [post]
func (h *MeasurementController) Pause(c *gin.Context) {

	id, _ := strconv.Atoi(c.Query("measurementId"))

	err := h.service.PauseMeasurement(uint(id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"paused": true})
}

// ResumeMeasurement godoc
// @Summary Возобновление измерения
// @Description Возобновляет измерение после паузы
// @Tags measurements
// @Produce json
// @Security BearerAuth
// @Param measurementId query int true "Measurement ID"
// @Success 200 {object} map[string]bool "resumed"
// @Failure 500 {object} map[string]string "error"
// @Router /measurements/resume [post]
func (h *MeasurementController) Resume(c *gin.Context) {

	id, _ := strconv.Atoi(c.Query("measurementId"))

	err := h.service.ResumeMeasurement(uint(id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"resumed": true})
}

// FinishMeasurement godoc
// @Summary Завершение измерения
// @Description Завершает измерение времени для шага
// @Tags measurements
// @Produce json
// @Security BearerAuth
// @Param measurementId query int true "Measurement ID"
// @Success 200 {object} map[string]bool "finished"
// @Failure 500 {object} map[string]string "error"
// @Router /measurements/finish [post]
func (h *MeasurementController) Finish(c *gin.Context) {

	id, _ := strconv.Atoi(c.Query("measurementId"))

	err := h.service.FinishMeasurement(uint(id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"finished": true})
}

// ResetMeasurement godoc
// @Summary Сброс измерений
// @Description Сбрасывает все измерения для шага
// @Tags measurements
// @Produce json
// @Security BearerAuth
// @Param stepId query int true "Step ID"
// @Success 200 {object} map[string]bool "reset"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /measurements/reset [post]
func (h *MeasurementController) Reset(c *gin.Context) {

	stepID, err := strconv.Atoi(c.Query("stepId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stepId"})
		return
	}

	err = h.service.ResetMeasurements(uint(stepID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"reset": true})
}

// DeleteMeasurement godoc
// @Summary Удаление замера
// @Description Удаляет замер времени по ID
// @Tags measurements
// @Produce json
// @Security BearerAuth
// @Param id path int true "Measurement ID"
// @Success 200 {object} map[string]bool "deleted"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /measurements/{id} [delete]
func (h *MeasurementController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid measurement id"})
		return
	}

	if err := h.service.DeleteMeasurement(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
