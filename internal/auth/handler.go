package auth

import (
	"api/config"
	"api/internal/models"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// --- VALIDATION HELPERS ---

func IsValidName(name string) bool {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < 3 {
		return false
	}
	// Ensure name is only letters and spaces (No Jay559!)
	for _, r := range trimmed {
		if !unicode.IsLetter(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func IsValidKenyanPhone(phone string) bool {
	// Pattern: 254 + 9 digits OR 07 + 8 digits OR 01 + 8 digits
	pattern := `^(254\d{9}|07\d{8}|01\d{8})$`
	match, _ := regexp.MatchString(pattern, phone)
	return match
}

func IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

// --- MAIN SIGNUP FUNCTION ---

func Register(c *gin.Context) {
	var input struct {
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
		Email     string `json:"email" binding:"required"`
		Phone     string `json:"phone" binding:"required"`
		Password  string `json:"password" binding:"required,min=6"`
		Role      string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	// 1. Validate First and Last Names
	if !IsValidName(input.FirstName) || !IsValidName(input.LastName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Names must be letters only (no numbers) and at least 2 characters"})
		return
	}

	// 2. Validate Email & Kenyan Phone
	if !IsValidEmail(input.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}
	if !IsValidKenyanPhone(input.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid phone. Use 254..., 07..., or 01..."})
		return
	}

	// 3. Prevent duplicate accounts
	var existingUser models.User
	if err := config.DB.Where("email = ? OR phone = ?", input.Email, input.Phone).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email or phone number already registered"})
		return
	}

	// 4. Hash Password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	// 5. Create User (MATCHING YOUR STRUCT FIELDS)
	user := models.User{
		ID:        uuid.New(),
		FirstName: input.FirstName, // Corrected
		LastName:  input.LastName,  // Corrected
		Email:     input.Email,
		Phone:     input.Phone,
		Password:  string(hashedPassword),
		Role:      strings.ToLower(input.Role),
		Status:    "active",
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Account created successfully!"})
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
