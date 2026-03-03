package controller

import (
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
// @Param user body models.User true "User data"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [post]
func (c *UserController) Create(ctx *gin.Context) {
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
