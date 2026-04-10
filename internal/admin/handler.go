package admin

import (
	"api/config"
	"api/internal/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListAllNannies - Admin sees all nannies and their verification profiles
func ListAllNannies(c *gin.Context) {
	// 1. Get pagination parameters from URL
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var nannies []models.User
	var total int64

	// 2. Count total records for the frontend to show "Total Pages"
	config.DB.Model(&models.User{}).Where("role = ?", "nanny").Count(&total)

	// 3. Fetch paginated data
	err := config.DB.Preload("NannyProfile").
		Where("role = ?", "nanny").
		Limit(limit).
		Offset(offset).
		Find(&nannies).Error

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch data"})
		return
	}

	c.JSON(200, gin.H{
		"data":      nannies,
		"total":     total,
		"page":      page,
		"last_page": math.Ceil(float64(total) / float64(limit)),
	})
}

// ListAllEmployers - Admin sees all employers and their residence details
func ListAllEmployers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var employers []models.User
	var total int64

	config.DB.Model(&models.User{}).Where("role = ?", "employer").Count(&total)

	config.DB.Preload("EmployerProfile").
		Where("role = ?", "employer").
		Limit(limit).Offset(offset).
		Find(&employers)

	c.JSON(200, gin.H{
		"data":      employers,
		"total":     total,
		"page":      page,
		"last_page": math.Ceil(float64(total) / float64(limit)),
	})
}

// ToggleUserStatus - Block or Unblock a user
func ToggleUserStatus(c *gin.Context) {
	var input struct {
		UserID string `json:"user_id" binding:"required"`
		Status string `json:"status" binding:"required"` // "active" or "blocked"
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := config.DB.Model(&models.User{}).Where("id = ?", input.UserID).Update("status", input.Status)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User status updated to " + input.Status})
}
