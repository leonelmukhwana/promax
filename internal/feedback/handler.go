package feedback

import (
	"api/config"
	"api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Nanny/Employer files a complaint (Invisible to users, Admin only)
func FileComplaint(c *gin.Context) {
	val, _ := c.Get("userID")
	reporterID, _ := uuid.Parse(val.(string))

	var input struct {
		TargetID    uuid.UUID `json:"target_id" binding:"required"`
		Category    string    `json:"category" binding:"required"`
		Description string    `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	complaint := models.Complaint{
		ReporterID:  reporterID,
		TargetID:    input.TargetID,
		Category:    input.Category,
		Description: input.Description,
	}

	config.DB.Create(&complaint)
	c.JSON(http.StatusCreated, gin.H{"message": "Report submitted securely to admin."})
}

// Employer rates a Nanny
func RateNanny(c *gin.Context) {
	val, _ := c.Get("userID")
	employerID, _ := uuid.Parse(val.(string))

	var rating models.Rating
	if err := c.ShouldBindJSON(&rating); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rating.EmployerID = employerID

	config.DB.Create(&rating)
	c.JSON(http.StatusCreated, gin.H{"message": "Rating submitted successfully"})
}

// Nanny views her own ratings
func GetMyRatings(c *gin.Context) {
	val, _ := c.Get("userID")
	var ratings []models.Rating
	config.DB.Where("nanny_id = ?", val).Find(&ratings)
	c.JSON(http.StatusOK, ratings)
}

// Admin views all complaints (The Safety Monitor)
func AdminListComplaints(c *gin.Context) {
	var complaints []models.Complaint
	config.DB.Find(&complaints)
	c.JSON(http.StatusOK, complaints)
}

// Admin views all ratings
func AdminListAllRatings(c *gin.Context) {
	var ratings []models.Rating
	config.DB.Find(&ratings)
	c.JSON(http.StatusOK, ratings)
}
