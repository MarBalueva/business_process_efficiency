package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"business_process_efficiency/internal/database"
	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/service"
)

type UserController struct {
	service *service.UserService
}

func NewUserController() *UserController {
	return &UserController{
		service: service.NewUserService(),
	}
}

// Create godoc
// @Summary Создать пользователя
// @Description Создает нового пользователя
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.CreateUserRequest true "User data"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [post]
func (c *UserController) Create(ctx *gin.Context) {

	var req models.CreateUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Login:      req.Login,
		Password:   req.Password,
		EmployeeID: req.EmployeeID,
	}

	err := c.service.CreateUser(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

// GetAll godoc
// @Summary Получить всех пользователей
// @Description Возвращает список пользователей
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.User
// @Failure 500 {object} map[string]string
// @Router /users [get]
func (c *UserController) GetAll(ctx *gin.Context) {
	var users []models.User

	err := database.DB.
		Preload("AccessGroups", "deleted_at IS NULL").
		Preload("AccessGroups.AccessGroup", "deleted_at IS NULL").
		Find(&users).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

// GetByID godoc
// @Summary Получить пользователя
// @Description Возвращает пользователя по ID
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 404 {object} map[string]string
// @Router /users/{id} [get]
func (c *UserController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")

	user, err := c.service.GetUserByID(parseUint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// Update godoc
// @Summary Обновить пользователя
// @Description Обновляет логин и пароль пользователя
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param user body models.UpdateUserRequest true "User data"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [put]
func (c *UserController) Update(ctx *gin.Context) {

	id := parseUint(ctx.Param("id"))

	var req models.UpdateUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.service.UpdateUser(id, req.Login, req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, _ := c.service.GetUserByID(id)

	ctx.JSON(http.StatusOK, user)
}

// Delete godoc
// @Summary Удалить пользователя
// @Description Удаляет пользователя (soft delete)
// @Tags users
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 204
// @Failure 500 {object} map[string]string
// @Router /users/{id} [delete]
func (c *UserController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	err := c.service.DeleteUser(parseUint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

type AccessGroupRequest struct {
	AccessGroupID uint `json:"access_group_id"`
}

// AddAccessGroup godoc
// @Summary Добавить группу доступа пользователю
// @Description Назначает пользователю группу доступа
// @Tags users
// @Accept json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param data body AccessGroupRequest true "Access group ID"
// @Success 201
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id}/access-groups [post]
func (c *UserController) AddAccessGroup(ctx *gin.Context) {
	id := ctx.Param("id")

	var req AccessGroupRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.service.AddAccessGroup(parseUint(id), req.AccessGroupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusCreated)
}

// RemoveAccessGroup godoc
// @Summary Удалить группу доступа пользователя
// @Description Удаляет группу доступа у пользователя
// @Tags users
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param group_id path int true "Access Group ID"
// @Success 204
// @Failure 500 {object} map[string]string
// @Router /users/{id}/access-groups/{group_id} [delete]
func (c *UserController) RemoveAccessGroup(ctx *gin.Context) {
	id := ctx.Param("id")
	groupID := ctx.Param("group_id")

	err := c.service.RemoveAccessGroup(parseUint(id), parseUint(groupID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// GetUserAccessGroups godoc
// @Summary Получить группы доступа пользователя
// @Description Возвращает список групп доступа пользователя
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {array} models.UserAccessGroup
// @Failure 500 {object} map[string]string
// @Router /users/{id}/access-groups [get]
func (c *UserController) GetUserAccessGroups(ctx *gin.Context) {

	idParam := ctx.Param("id")

	var userID uint
	if _, err := fmt.Sscan(idParam, &userID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	groups, err := c.service.GetUserAccessGroups(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, groups)
}

// GetByEmployee godoc
// @Summary Получить пользователя по сотруднику
// @Description Возвращает пользователя по ID сотрудника
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param employeeId path int true "Employee ID"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /users/by-employee/{employeeId} [get]
func (c *UserController) GetByEmployee(ctx *gin.Context) {

	employeeID := parseUint(ctx.Param("employeeId"))

	user, err := c.service.GetUserByEmployeeID(employeeID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}
