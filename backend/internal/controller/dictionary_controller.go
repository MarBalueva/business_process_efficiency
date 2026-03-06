package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/service"
)

type DictionaryController struct {
	service *service.DictionaryService
}

func NewDictionaryController(service *service.DictionaryService) *DictionaryController {
	return &DictionaryController{service: service}
}

func (h *DictionaryController) List(c *gin.Context, model interface{}) {

	if err := h.service.List(model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка получения справочника",
		})
		return
	}

	c.JSON(http.StatusOK, model)
}

func (h *DictionaryController) Create(c *gin.Context, model interface{}) {

	var input models.DictionaryRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверные данные",
		})
		return
	}

	if err := h.service.Create(model, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Запись создана",
	})
}

func (h *DictionaryController) Update(c *gin.Context, model interface{}) {

	id, _ := strconv.Atoi(c.Param("id"))

	var input models.DictionaryRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Неверные данные",
		})
		return
	}

	if err := h.service.Update(model, uint(id), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Запись обновлена",
	})
}

func (h *DictionaryController) Delete(c *gin.Context, model interface{}) {

	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.Delete(model, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка удаления",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Запись удалена",
	})
}

// GetAllDictionaries godoc
// @Summary Получить все справочники
// @Description Возвращает все справочники системы
// @Tags dictionary
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.DictionariesResponse
// @Router /dict [get]
func (h *DictionaryController) GetAll(c *gin.Context) {

	data, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка получения справочников",
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// ListDepartments godoc
// @Summary Получить список отделов
// @Description Возвращает список отделов (без удалённых)
// @Tags dictionary
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.DepartmentResponse
// @Failure 500 {object} map[string]string
// @Router /dict/departments [get]
func (h *DictionaryController) ListDepartments(c *gin.Context) {
	var items []models.Department
	h.List(c, &items)
}

// CreateDepartment godoc
// @Summary Создать отдел
// @Description Добавляет новый отдел
// @Tags dictionary
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body models.DictionaryRequest true "Данные отдела"
// @Success 200 {object} models.DepartmentResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /dict/departments [post]
func (h *DictionaryController) CreateDepartment(c *gin.Context) {
	h.Create(c, &models.Department{})
}

// UpdateDepartment godoc
// @Summary Обновить отдел
// @Description Обновляет существующий отдел
// @Tags dictionary
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID отдела"
// @Param data body models.DictionaryRequest true "Новые данные"
// @Success 200 {object} models.DepartmentResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /dict/departments/{id} [put]
func (h *DictionaryController) UpdateDepartment(c *gin.Context) {
	h.Update(c, &models.Department{})
}

// DeleteDepartment godoc
// @Summary Удалить отдел
// @Description Помечает отдел как удалённый (soft delete)
// @Tags dictionary
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID отдела"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /dict/departments/{id} [delete]
func (h *DictionaryController) DeleteDepartment(c *gin.Context) {
	h.Delete(c, &models.Department{})
}

// ListPositions godoc
// @Summary Получить список должностей
// @Description Возвращает список должностей
// @Tags dictionary
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.PositionResponse
// @Router /dict/positions [get]
func (h *DictionaryController) ListPositions(c *gin.Context) {
	var items []models.Position
	h.List(c, &items)
}

// CreatePosition godoc
// @Summary Создать должность
// @Description Добавляет новую должность
// @Tags dictionary
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body models.DictionaryRequest true "Данные"
// @Success 200 {object} models.PositionResponse
// @Router /dict/positions [post]
func (h *DictionaryController) CreatePosition(c *gin.Context) {
	h.Create(c, &models.Position{})
}

// UpdatePosition godoc
// @Summary Обновить должность
// @Description Обновляет должность
// @Tags dictionary
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID должности"
// @Param data body models.DictionaryRequest true "Новые данные"
// @Success 200 {object} models.PositionResponse
// @Router /dict/positions/{id} [put]
func (h *DictionaryController) UpdatePosition(c *gin.Context) {
	h.Update(c, &models.Position{})
}

// DeletePosition godoc
// @Summary Удалить должность
// @Description Помечает должность удалённой
// @Tags dictionary
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID должности"
// @Success 200 {object} map[string]string
// @Router /dict/positions/{id} [delete]
func (h *DictionaryController) DeletePosition(c *gin.Context) {
	h.Delete(c, &models.Position{})
}

// ListAccessGroups godoc
// @Summary Получить список групп доступа
// @Description Возвращает группы доступа
// @Tags dictionary
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.AccessGroupResponse
// @Router /dict/access_groups [get]
func (h *DictionaryController) ListAccessGroups(c *gin.Context) {
	var items []models.AccessGroup
	h.List(c, &items)
}

// CreateAccessGroup godoc
// @Summary Создать группу доступа
// @Description Добавляет группу доступа
// @Tags dictionary
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body models.DictionaryRequest true "Данные"
// @Success 200 {object} models.AccessGroupResponse
// @Router /dict/access_groups [post]
func (h *DictionaryController) CreateAccessGroup(c *gin.Context) {
	h.Create(c, &models.AccessGroup{})
}

// UpdateAccessGroup godoc
// @Summary Обновить группу доступа
// @Description Обновляет группу доступа
// @Tags dictionary
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID группы"
// @Param data body models.DictionaryRequest true "Новые данные"
// @Success 200 {object} models.AccessGroupResponse
// @Router /dict/access_groups/{id} [put]
func (h *DictionaryController) UpdateAccessGroup(c *gin.Context) {
	h.Update(c, &models.AccessGroup{})
}

// DeleteAccessGroup godoc
// @Summary Удалить группу доступа
// @Description Помечает группу доступа удалённой
// @Tags dictionary
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID группы"
// @Success 200 {object} map[string]string
// @Router /dict/access_groups/{id} [delete]
func (h *DictionaryController) DeleteAccessGroup(c *gin.Context) {
	h.Delete(c, &models.AccessGroup{})
}
