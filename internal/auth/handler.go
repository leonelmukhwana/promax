package auth

import (
	"api/config"
	"api/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 1. REGISTER (With Role Protection)
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// SECURITY: Hard-block anyone trying to register as Admin
	if input.Role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized role selection."})
		return
	}

	hashedPassword, _ := HashPassword(input.Password)
	user := models.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Phone:     input.Phone,
		Password:  hashedPassword,
		Role:      input.Role, // Will be 'nanny' or 'employer'
		Status:    "inactive",
	}

	tx := config.DB.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "User with this Email/Phone already exists"})
		return
	}

	// Create OTP
	code := GenerateOTP()
	otp := models.OTP{
		UserID:    user.ID,
		Code:      code,
		Purpose:   "verification",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	tx.Create(&otp)

	// Audit Log the Registration
	newVal, _ := json.Marshal(user)
	audit := models.AuditLog{
		UserID:    user.ID,
		Action:    "USER_REGISTER",
		Resource:  "Users",
		NewValue:  string(newVal),
		IPAddress: c.ClientIP(),
	}
	tx.Create(&audit)

	tx.Commit()

	// Instant Background Email
	go SendOTPEmail(user.Email, code)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. Verify your email to activate account.",
		"role":    user.Role,
	})
}

// 2. LOGIN (With Status & Role check)
func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check Password
	if !CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Security: Check if blocked
	if user.Status == "blocked" {
		c.JSON(http.StatusForbidden, gin.H{"error": "This account has been suspended. Please contact admin."})
		return
	}

	// Security: Check if verified
	if !user.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Please verify your email before logging in."})
		return
	}

	// Generate Token with Role included in Claims
	token, err := GenerateToken(user.ID.String(), user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"role":  user.Role,
		"user": gin.H{
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
		},
	})
}

// 2. FORGOT PASSWORD (Request OTP)
func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "If email exists, OTP has been sent"}) // Security: don't reveal email existence
		return
	}

	code := GenerateOTP()
	config.DB.Create(&models.OTP{UserID: user.ID, Code: code, Purpose: "reset", ExpiresAt: time.Now().Add(10 * time.Minute)})
	go SendOTPEmail(user.Email, code)

	c.JSON(http.StatusOK, gin.H{"message": "Reset OTP sent"})
}

// 3. RESET PASSWORD (Use OTP to set new password)
func ResetPassword(c *gin.Context) {
	var input struct {
		Email       string `json:"email" binding:"required"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	config.DB.Where("email = ?", input.Email).First(&user)

	var otp models.OTP
	err := config.DB.Where("user_id = ? AND code = ? AND purpose = 'reset' AND expires_at > ?", user.ID, input.Code, time.Now()).First(&otp).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	newHash, _ := HashPassword(input.NewPassword)
	config.DB.Model(&user).Update("password", newHash)
	config.DB.Delete(&otp)

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

// MUST be Capital V to be visible in routes.go
func VerifyOTP(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	// Use config.DB (the exported global DB)
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var otp models.OTP
	// Verify code exists, belongs to user, and hasn't expired
	err := config.DB.Where("user_id = ? AND code = ? AND expires_at > ?", user.ID, input.Code, time.Now()).First(&otp).Error

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	// Update User to active and verified
	config.DB.Model(&user).Updates(map[string]interface{}{
		"is_verified": true,
		"status":      "active",
	})

	// Delete the OTP so it can't be reused
	config.DB.Delete(&otp)

	c.JSON(http.StatusOK, gin.H{"message": "Account verified successfully. You can now login."})
}

// logout functionality
func Logout(c *gin.Context) {
	// In the future, you can add the token to a Redis blacklist here
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// change password functionality
func ChangePassword(c *gin.Context) {
	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	// 1. Get user ID from JWT context (set by middleware)
	userID := c.MustGet("userID").(string)

	var user models.User
	config.DB.First(&user, "id = ?", userID)

	// 2. Verify Old Password
	if !CheckPasswordHash(input.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password incorrect"})
		return
	}

	// 3. Hash and Save New Password
	hashed, _ := HashPassword(input.NewPassword)
	config.DB.Model(&user).Update("password", hashed)

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// resend OTP again
func ResendOTP(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// 1. Delete old OTPs
	config.DB.Where("user_id = ?", user.ID).Delete(&models.OTP{})

	// 2. Generate new OTP
	code := GenerateOTP()
	newOtp := models.OTP{
		UserID:    user.ID,
		Code:      code,
		Purpose:   "verification",
		ExpiresAt: time.Now().Add(time.Minute * 15),
	}
	config.DB.Create(&newOtp)

	// 3. Print to console for your testing
	fmt.Printf("\n🔄 [RESEND] New OTP for %s: %s\n", user.Email, code)
	go SendOTPEmail(user.Email, code)

	c.JSON(200, gin.H{"message": "New verification code sent"})
}
