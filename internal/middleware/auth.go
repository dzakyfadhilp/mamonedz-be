package middleware

import (
	"strings"

	"mamonedz/internal/services"
	"mamonedz/pkg/response"

	"github.com/gin-gonic/gin"
)

func Auth(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, 401, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		userID, err := authService.ValidateToken(tokenString)
		if err != nil {
			response.Error(c, 401, "Invalid or expired token")
			c.Abort()
			return
		}

		user, err := authService.GetUserByID(*userID)
		if err != nil {
			response.Error(c, 401, "User not found")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("user", user)
		c.Next()
	}
}
