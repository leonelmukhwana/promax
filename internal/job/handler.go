package job

import (
	"api/config"
	"api/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NannyJobView - DTO to ensure Nannies NEVER see the salary amount
type NannyJobView struct {
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Requirements string `json:"requirements"`
	Location     string `json:"location"`
	Status       string `json:"status"`
}

// CreateJob (Employer)
func CreateJob(c *gin.Context) {
	val, _ := c.Get("userID")
	employerID, _ := uuid.Parse(val.(string))

	var job models.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job.EmployerID = employerID
	job.Status = "open"

	if err := config.DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not post job"})
		return
	}

	c.JSON(http.StatusCreated, job)
}

// AdminAssign (Admin Only) - Link a Nanny to a Job
func AdminAssign(c *gin.Context) {
	var input struct {
		JobID   uint      `json:"job_id" binding:"required"`
		NannyID uuid.UUID `json:"nanny_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the job with the NannyID and change status to assigned
	result := config.DB.Model(&models.Job{}).Where("id = ?", input.JobID).
		Updates(map[string]interface{}{
			"nanny_id": input.NannyID,
			"status":   "assigned",
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Assignment failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nanny successfully assigned to job"})
}

// GetNannyJobs (Nanny) - Restricted View
func GetNannyJobs(c *gin.Context) {
	val, _ := c.Get("userID")
	nannyID, _ := uuid.Parse(val.(string))

	var jobs []models.Job
	config.DB.Where("nanny_id = ?", nannyID).Find(&jobs)

	// Map to the View DTO to exclude Salary
	var response []NannyJobView
	for _, j := range jobs {
		response = append(response, NannyJobView{
			ID:           j.ID,
			Title:        j.Title,
			Description:  j.Description,
			Requirements: j.Requirements,
			Location:     j.Location,
			Status:       j.Status,
		})
	}

	c.JSON(http.StatusOK, response)
}

// AdminListJobs (Admin) - Monitor Table
func AdminListJobs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	offset := (page - 1) * limit

	var jobs []models.Job
	var total int64

	config.DB.Model(&models.Job{}).Count(&total)

	config.DB.Preload("Employer").Preload("Nanny").
		Limit(limit).Offset(offset).
		Order("created_at DESC"). // Newest jobs first
		Find(&jobs)

	c.JSON(200, gin.H{
		"data":  jobs,
		"total": total,
		"page":  page,
	})
}
