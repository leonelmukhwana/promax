package middleware

import (
	"api/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

// RoleRequired checks if the logged-in user has the specific required role
func RoleRequired(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from the AuthRequired middleware context
		val, exists := c.Get("currentUser")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		user := val.(models.User)

		if user.Role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: You do not have the required permissions.",
			})
			return
		}

		c.Next()
	}
}
