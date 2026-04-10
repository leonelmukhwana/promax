package middleware

import (
	"net/http"
	"os"
	"strings"

	"api/config"
	"api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Extract userID from claims (sub)
		sub, ok := claims["sub"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID missing from token"})
			return
		}

		// Check if user is blocked or inactive
		var user models.User
		if err := config.DB.First(&user, "id = ?", sub).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		if user.Status == "blocked" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Your account is blocked"})
			return
		}

		// --- THE FIX: SET THE KEYS YOUR HANDLERS EXPECT ---

		// 1. Set userID as a string (Fixes your panic)
		c.Set("userID", sub)

		// 2. Set userRole from token (Useful for role-based logic)
		c.Set("userRole", user.Role)

		// 3. Keep the full user object just in case
		c.Set("currentUser", user)

		c.Next()
	}
}

// AdminOnly ensures that only users with the 'admin' role can proceed
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the role we set in AuthRequired
		role, exists := c.Get("userRole")

		if !exists || role != "admin" {
			// If the role isn't 'admin', block the request immediately
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: This action requires Administrator privileges",
			})
			return
		}

		// If they are admin, let them through
		c.Next()
	}
}
