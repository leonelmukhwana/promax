package routes

import (
	"api/internal/admin"
	"api/internal/auth"
	"api/internal/booking"
	"api/internal/feedback"
	"api/internal/job"
	"api/internal/middleware"
	"api/internal/profile"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	v1 := r.Group("/api/v1")

	// --- 1. PUBLIC AUTH ROUTES ---
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", auth.Register) // Changed to SignUp to match our new logic
		authGroup.POST("/login", auth.Login)
		authGroup.POST("/verify-otp", auth.VerifyOTP)
		authGroup.POST("/resend-otp", middleware.RateLimit(3), auth.ResendOTP)
		authGroup.POST("/forgot-password", auth.ForgotPassword)
		authGroup.POST("/reset-password", auth.ResetPassword)
	}

	// --- 2. PROTECTED ROUTES (Requires Login) ---
	protected := v1.Group("/")
	protected.Use(middleware.AuthRequired())
	{
		// Profiles
		protected.POST("/profile/nanny", profile.CreateOrUpdateNannyProfile)
		protected.POST("/profile/employer", profile.CreateOrUpdateEmployerProfile)
		protected.GET("/profile/me", profile.GetMyProfile)
		protected.DELETE("/profile", profile.DeleteProfile)

		// Jobs & Bookings
		protected.POST("/jobs", job.CreateJob)
		protected.GET("/jobs/assigned", job.GetNannyJobs)
		protected.POST("/bookings", booking.CreateBooking)
		protected.GET("/bookings/me", booking.GetMyBooking)
		protected.PUT("/bookings/:id", booking.UpdateBooking)

		// Feedback
		protected.POST("/complaints", feedback.FileComplaint)
		protected.POST("/ratings", feedback.RateNanny)
		protected.GET("/ratings/me", feedback.GetMyRatings)

		// --- 3. ADMIN ONLY (Double Protected) ---
		adminGroup := protected.Group("/admin")
		adminGroup.Use(middleware.AdminOnly())
		{
			// Admin Monitoring
			adminGroup.GET("/jobs", job.AdminListJobs)
			adminGroup.GET("/complaints", feedback.AdminListComplaints)
			adminGroup.GET("/ratings", feedback.AdminListAllRatings)
			adminGroup.DELETE("/bookings/:id", booking.AdminDelete)
			adminGroup.POST("/jobs/assign", job.AdminAssign)

			// User Management
			adminGroup.GET("/nannies", admin.ListAllNannies)
			adminGroup.GET("/employers", admin.ListAllEmployers)
			adminGroup.PATCH("/users/status", admin.ToggleUserStatus)
		}
	}
}
