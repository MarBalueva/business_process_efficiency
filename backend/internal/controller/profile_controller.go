package controller

import (
	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	service *service.ProfileService
}

func NewProfileController(service *service.ProfileService) *ProfileController {
	return &ProfileController{service: service}
}

// GetProfile godoc
// @Summary Получить профиль текущего пользователя
// @Description Возвращает логин пользователя и данные сотрудника (ФИО).
// @Tags profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Профиль пользователя"
// @Failure 401 {object} map[string]string "Неавторизован"
// @Failure 404 {object} map[string]string "Пользователь или сотрудник не найден"
// @Router /profile/me [get]
func (c *ProfileController) GetProfile(ctx *gin.Context) {
	userIDRaw, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDRaw.(uint)

	user, employee, err := c.service.GetProfile(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user or employee not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"login": user.Login,
		"employee": gin.H{
			"last_name":   employee.LastName,
			"first_name":  employee.FirstName,
			"middle_name": employee.MiddleName,
		},
	})
}

// UpdateProfile godoc
// @Summary Обновить профиль текущего пользователя
// @Description Позволяет изменить пароль пользователя и/или ФИО сотрудника. Передавать можно только поля, которые нужно обновить.
// @Tags profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param profile body models.ProfileUpdateRequest true "Данные для обновления профиля"
// @Success 200 {object} map[string]string "Профиль успешно обновлён"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 401 {object} map[string]string "Неавторизован"
// @Failure 404 {object} map[string]string "Пользователь или сотрудник не найден"
// @Failure 500 {object} map[string]string "Ошибка при обновлении профиля"
// @Router /profile/me [put]
func (c *ProfileController) UpdateProfile(ctx *gin.Context) {
	userIDRaw, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDRaw.(uint)

	var input models.ProfileUpdateRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if err := c.service.UpdateProfile(userID, input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}
