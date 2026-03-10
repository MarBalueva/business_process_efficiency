package controller

import (
	"net/http"
	"strconv"

	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/service"

	"github.com/gin-gonic/gin"
)

type ProcessController struct {
	service *service.ProcessService
}

func NewProcessController(service *service.ProcessService) *ProcessController {
	return &ProcessController{service: service}
}

// GetRegistry godoc
// @Summary Получение дерева процессов
// @Description Возвращает все папки и процессы в виде дерева
// @Tags processes
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ProcessRegistryFolder
// @Failure 500 {object} map[string]string "error"
// @Router /processes/registry [get]
func (h *ProcessController) GetRegistry(c *gin.Context) {

	data, err := h.service.GetRegistryTree()

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetProcess godoc
// @Summary Получение процесса по ID
// @Description Возвращает данные процесса с версиями и шагами
// @Tags processes
// @Produce json
// @Security BearerAuth
// @Param id path int true "Process ID"
// @Success 200 {object} models.Process
// @Failure 404 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /processes/{id} [get]
func (h *ProcessController) GetProcess(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))

	process, err := h.service.GetProcess(uint(id))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "process not found"})
		return
	}

	c.JSON(http.StatusOK, process)
}

// CreateProcess godoc
// @Summary Создание процесса
// @Description Создает новый процесс в папке
// @Tags processes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param process body models.CreateProcessRequest true "Process data"
// @Success 201 {object} models.Process
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /processes [post]
func (h *ProcessController) CreateProcess(c *gin.Context) {
	var req models.CreateProcessRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	process, err := h.service.CreateProcess(req.Name, req.FolderID, userID, req.RegularityCount, req.RegularityUnit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, process)
}

// UpdateProcess godoc
// @Summary Обновление процесса
// @Description Обновляет данные процесса по ID
// @Tags processes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Process ID"
// @Param process body models.UpdateProcessRequest true "Updated process data"
// @Success 200 {object} models.Process
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /processes/{id} [put]
func (h *ProcessController) UpdateProcess(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req models.UpdateProcessRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	process, err := h.service.UpdateProcess(uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, process)
}

// DeleteProcess godoc
// @Summary Удаление процесса
// @Description Удаляет процесс по ID
// @Tags processes
// @Produce json
// @Security BearerAuth
// @Param id path int true "Process ID"
// @Success 200 {object} map[string]bool "deleted"
// @Failure 500 {object} map[string]string "error"
// @Router /processes/{id} [delete]
func (h *ProcessController) DeleteProcess(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))

	err := h.service.DeleteProcess(uint(id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// DeleteVersion godoc
// @Summary Удаление версии процесса
// @Description Удаляет версию процесса по ID
// @Tags process-versions
// @Produce json
// @Security BearerAuth
// @Param id path int true "Version ID"
// @Success 200 {object} map[string]bool "deleted"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /processes/versions/{id} [delete]
func (h *ProcessController) DeleteVersion(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.service.DeleteVersion(uint(id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// CreateVersion godoc
// @Summary Создание версии процесса
// @Description Создает новую версию процесса по ID
// @Tags process-versions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param version body models.CreateVersionRequest true "Version data"
// @Success 201 {object} models.ProcessVersion
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /processes/versions [post]
func (h *ProcessController) CreateVersion(c *gin.Context) {

	type request struct {
		ProcessID uint `json:"processId"`
	}

	var req request

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	version, err := h.service.CreateVersion(req.ProcessID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, version)
}

// CreateFolder godoc
// @Summary Создание папки процесса
// @Description Создает новую папку для процессов
// @Tags process-folders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param folder body models.CreateFolderRequest true "Folder data"
// @Success 201 {object} models.ProcessFolder
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /process-folders [post]
func (h *ProcessController) CreateFolder(c *gin.Context) {
	type request struct {
		Name     string `json:"name"`
		ParentID *uint  `json:"parentId"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.service.CreateFolder(req.Name, req.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

// UpdateFolder godoc
// @Summary Обновление папки процесса
// @Description Обновляет название и родительскую папку по ID
// @Tags process-folders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Folder ID"
// @Param folder body models.CreateFolderRequest true "Folder data"
// @Success 200 {object} map[string]bool "updated"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /process-folders/{id} [put]
func (h *ProcessController) UpdateFolder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	type request struct {
		Name     string `json:"name"`
		ParentID *uint  `json:"parentId"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateFolder(uint(id), req.Name, req.ParentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": true})
}

// DeleteFolder godoc
// @Summary Удаление папки процесса
// @Description Удаляет папку процесса по ID
// @Tags process-folders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Folder ID"
// @Success 200 {object} map[string]bool "deleted"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /process-folders/{id} [delete]
func (h *ProcessController) DeleteFolder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.service.DeleteFolder(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// MoveProcess godoc
// @Summary Перемещение процесса в папку
// @Description Изменяет папку процесса по ID
// @Tags processes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Process ID"
// @Param move body map[string]interface{} true "Move payload"
// @Success 200 {object} map[string]bool "moved"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /processes/{id}/move [patch]
func (h *ProcessController) MoveProcess(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid process id"})
		return
	}

	var req struct {
		FolderID *uint `json:"folderId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.MoveProcess(uint(id), req.FolderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"moved": true})
}

// MoveFolder godoc
// @Summary Перемещение папки
// @Description Изменяет родительскую папку для папки по ID
// @Tags process-folders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Folder ID"
// @Param move body map[string]interface{} true "Move payload"
// @Success 200 {object} map[string]bool "moved"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /process-folders/{id}/move [patch]
func (h *ProcessController) MoveFolder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder id"})
		return
	}

	var req struct {
		ParentID *uint `json:"parentId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.MoveFolder(uint(id), req.ParentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"moved": true})
}
