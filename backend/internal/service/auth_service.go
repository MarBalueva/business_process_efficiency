package service

import (
	"time"

	"business_process_efficiency/internal/database"
	"business_process_efficiency/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	jwtSecret string
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{jwtSecret: secret}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *AuthService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	var accessGroups []models.UserAccessGroup

	err := database.DB.
		Preload("AccessGroup", "deleted_at IS NULL").
		Where("user_id = ?", user.ID).
		Find(&accessGroups).Error
	if err != nil {
		return "", err
	}

	codes := []string{}
	for _, ag := range accessGroups {
		if ag.AccessGroup.Code != "" {
			codes = append(codes, ag.AccessGroup.Code)
		}
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"login":   user.Login,
		"groups":  codes,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ParseToken(tokenStr string) (*jwt.Token, jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	return token, claims, err
}

func (s *AuthService) HasAccess(claims jwt.MapClaims, code string) bool {
	if groups, ok := claims["groups"].([]interface{}); ok {
		for _, g := range groups {
			if g.(string) == code {
				return true
			}
		}
	}
	return false
}
