package middleware

import (
	"net/http"
	"personal-finance-tracker/internal/models"
	"personal-finance-tracker/internal/service"
	"personal-finance-tracker/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware проверяет JWT (подпись и срок) и валидность сессии в БД
func AuthMiddleware(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Формат: "Bearer {token}"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Сначала валидируем подпись JWT, затем сверяем сессию в БД
		_, jwtErr := utils.ValidateJWT(token)
		if jwtErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		user, err := userService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Сохраняем пользователя в контекст
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Next()
	}
}

func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userObj, ok := user.(*models.User)
	return userObj, ok
}
