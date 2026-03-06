package controller

import (
	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EmployeeController struct {
	Service *service.EmployeeService
}

func NewEmployeeController(service *service.EmployeeService) *EmployeeController {
	return &EmployeeController{Service: service}
}

// CreateEmployee godoc
// @Summary Создание нового сотрудника
// @Description Создает сотрудника с данными из запроса
// @Tags employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param employee body models.EmployeeCreateRequest true "Employee data"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /employees [post]
func (c *EmployeeController) CreateEmployee(ctx *gin.Context) {
	var req models.EmployeeCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
		return
	}

	if err := c.Service.Create(req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Сотрудник создан"})
}

// GetAllEmployees godoc
// @Summary Получение всех сотрудников
// @Description Возвращает список всех сотрудников (неудаленных)
// @Tags employees
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.EmployeeFullResponse
// @Failure 500 {object} map[string]string "error"
// @Router /employees [get]
func (c *EmployeeController) GetAllEmployees(ctx *gin.Context) {
	employees, err := c.Service.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, employees)
}

// GetEmployee godoc
// @Summary Получение сотрудника по ID
// @Description Возвращает данные сотрудника по ID
// @Tags employees
// @Produce json
// @Security BearerAuth
// @Param id path int true "Employee ID"
// @Success 200 {object} models.EmployeeFullResponse
// @Failure 404 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /employees/{id} [get]
func (c *EmployeeController) GetEmployee(ctx *gin.Context) {
	id := parseUint(ctx.Param("id"))
	employee, err := c.Service.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Сотрудник не найден"})
		return
	}

	ctx.JSON(http.StatusOK, employee)
}

// UpdateEmployee godoc
// @Summary Обновление данных сотрудника
// @Description Обновляет данные сотрудника по ID
// @Tags employees
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Employee ID"
// @Param employee body models.EmployeeCreateRequest true "Employee data"
// @Success 200 {object} map[string]string "message"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /employees/{id} [put]
func (c *EmployeeController) UpdateEmployee(ctx *gin.Context) {
	id := parseUint(ctx.Param("id"))
	var req models.EmployeeCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
		return
	}

	if err := c.Service.Update(id, req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Сотрудник обновлен"})
}

// DeleteEmployee godoc
// @Summary Удаление сотрудника
// @Description Помечает сотрудника как удаленного (deleted_at)
// @Tags employees
// @Produce json
// @Security BearerAuth
// @Param id path int true "Employee ID"
// @Success 200 {object} map[string]string "message"
// @Failure 500 {object} map[string]string "error"
// @Router /employees/{id} [delete]
func (c *EmployeeController) DeleteEmployee(ctx *gin.Context) {
	id := parseUint(ctx.Param("id"))
	if err := c.Service.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Сотрудник удален"})
}

func parseUint(s string) uint {
	id, _ := strconv.ParseUint(s, 10, 64)
	return uint(id)
}
