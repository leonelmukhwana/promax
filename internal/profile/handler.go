package profile

import (
	"api/config"
	"api/internal/models"
	"api/internal/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// --- NANNY PROFILE CRUD ---

func CreateOrUpdateNannyProfile(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	userUUID, _ := uuid.Parse(userID)

	// 1. Parse Form Data
	homeCounty := c.PostForm("home_county")
	gender := c.PostForm("gender")
	dobStr := c.PostForm("dob") // Format: 1995-01-02
	location := c.PostForm("location")

	dob, _ := time.Parse("2006-01-02", dobStr)
	age := utils.CalculateAge(dob)

	// 2. Handle File Uploads (Selfie & ID)
	selfie, _ := c.FormFile("selfie")
	idCard, _ := c.FormFile("id_card")

	selfiePath := ""
	if selfie != nil {
		selfiePath = fmt.Sprintf("uploads/selfies/%s_%s", userID, selfie.Filename)
		c.SaveUploadedFile(selfie, selfiePath)
	}

	idPath := ""
	if idCard != nil {
		idPath = fmt.Sprintf("uploads/ids/%s_%s", userID, idCard.Filename)
		c.SaveUploadedFile(idCard, idPath)
	}

	profile := models.NannyProfile{
		UserID:      userUUID,
		HomeCounty:  homeCounty,
		Gender:      gender,
		DateOfBirth: dob,
		Age:         age,
		SelfieURL:   selfiePath,
		IDCardURL:   idPath,
		Location:    location,
	}

	// Upsert (Update or Insert)
	err := config.DB.Where(models.NannyProfile{UserID: userUUID}).
		Assign(profile).
		FirstOrCreate(&profile).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nanny profile saved successfully", "profile": profile})
}

// --- EMPLOYER PROFILE CRUD ---

func CreateOrUpdateEmployerProfile(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	userUUID, _ := uuid.Parse(userID)

	var input models.EmployerProfile
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.UserID = userUUID

	err := config.DB.Where(models.EmployerProfile{UserID: userUUID}).
		Assign(input).
		FirstOrCreate(&input).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employer profile saved successfully", "profile": input})
}

// --- GENERAL READ & DELETE ---

// --- READ (Get My Profile) ---
func GetMyProfile(c *gin.Context) {
	val, _ := c.Get("userID")
	role, _ := c.Get("userRole") // Assuming role is set in middleware
	userID := val.(string)

	if role == "nanny" {
		var profile models.NannyProfile
		if err := config.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusOK, profile)
	} else {
		var profile models.EmployerProfile
		if err := config.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusOK, profile)
	}
}

// --- DELETE (Soft Delete) ---
func DeleteProfile(c *gin.Context) {
	val, _ := c.Get("userID")
	role, _ := c.Get("userRole")
	userID := val.(string)

	if role == "nanny" {
		if err := config.DB.Where("user_id = ?", userID).Delete(&models.NannyProfile{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
			return
		}
	} else {
		if err := config.DB.Where("user_id = ?", userID).Delete(&models.EmployerProfile{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile soft-deleted successfully"})
}
