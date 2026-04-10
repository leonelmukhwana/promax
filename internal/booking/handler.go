package booking

import (
	"api/config"
	"api/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateBooking (Nanny) - IDEMPOTENT
func CreateBooking(c *gin.Context) {
	val, _ := c.Get("userID")
	nannyID, _ := uuid.Parse(val.(string))

	var input struct {
		Date      string `json:"date" binding:"required"`       // 2026-04-12
		StartTime string `json:"start_time" binding:"required"` // 09:00
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. IDEMPOTENCY CHECK: Is there already a pending booking?
	var existing models.Booking
	if err := config.DB.Where("nanny_id = ? AND status = ?", nannyID, "pending").First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You already have a pending interview booking."})
		return
	}

	// 2. Parse time and enforce 5PM limit
	date, _ := time.Parse("2006-01-02", input.Date)
	start, _ := time.Parse("15:04", input.StartTime)
	limit, _ := time.Parse("15:04", "17:00")

	if !start.Before(limit) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Interviews must be before 5:00 PM."})
		return
	}

	// 3. Create
	newBooking := models.Booking{
		NannyID:    nannyID,
		BookedDate: date,
		StartTime:  start,
		EndTime:    start.Add(30 * time.Minute),
		Status:     "pending",
	}

	if err := config.DB.Create(&newBooking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Booking failed"})
		return
	}

	c.JSON(http.StatusCreated, newBooking)
}

// GetMyBooking (Nanny)
func GetMyBooking(c *gin.Context) {
	val, _ := c.Get("userID")
	var booking models.Booking
	if err := config.DB.Where("nanny_id = ?", val).First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No booking found"})
		return
	}
	c.JSON(http.StatusOK, booking)
}

// UpdateBooking (Nanny)
func UpdateBooking(c *gin.Context) {
	id := c.Param("id")
	val, _ := c.Get("userID")

	var input struct {
		Date      string `json:"date"`
		StartTime string `json:"start_time"`
	}
	c.ShouldBindJSON(&input)

	var booking models.Booking
	if err := config.DB.Where("id = ? AND nanny_id = ?", id, val).First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	if input.Date != "" {
		booking.BookedDate, _ = time.Parse("2006-01-02", input.Date)
	}
	if input.StartTime != "" {
		start, _ := time.Parse("15:04", input.StartTime)
		booking.StartTime = start
		booking.EndTime = start.Add(30 * time.Minute)
	}

	config.DB.Save(&booking)
	c.JSON(http.StatusOK, gin.H{"message": "Booking updated successfully"})
}

// AdminListAll (Admin)
func AdminListAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	var bookings []models.Booking
	var total int64

	config.DB.Model(&models.Booking{}).Count(&total)

	config.DB.Preload("Nanny").
		Limit(limit).Offset(offset).
		Order("booked_date ASC"). // Show upcoming interviews first
		Find(&bookings)

	c.JSON(200, gin.H{
		"data":  bookings,
		"total": total,
		"page":  page,
	})
}

// AdminDelete (Admin) - Force re-booking
func AdminDelete(c *gin.Context) {
	id := c.Param("id")
	// Use Unscoped to bypass Soft Delete so the nanny can book again immediately
	if err := config.DB.Unscoped().Delete(&models.Booking{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Booking deleted. Nanny can now book a new slot."})
}
