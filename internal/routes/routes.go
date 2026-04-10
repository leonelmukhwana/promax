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
	{
		authGroup := v1.Group("/auth")
		{
			// Public Routes
			authGroup.POST("/register", auth.Register)
			authGroup.POST("/login", auth.Login)
			authGroup.POST("/verify-otp", auth.VerifyOTP)

			// 4. Resend OTP (Rate Limited: 3 requests per minute)
			authGroup.POST("/resend-otp", middleware.RateLimit(3), auth.ResendOTP)

			authGroup.POST("/forgot-password", auth.ForgotPassword)
			authGroup.POST("/reset-password", auth.ResetPassword)

			// 1 & 2. Protected Routes (Requires Login)
			protected := authGroup.Group("/")
			protected.Use(middleware.AuthRequired())
			{
				protected.POST("/change-password", auth.ChangePassword)
				protected.POST("/logout", auth.Logout)
			}
		}

		// Inside SetupRoutes...
		protected := v1.Group("/")
		protected.Use(middleware.AuthRequired())
		{
			// Nanny Routes
			protected.POST("/profile/nanny", profile.CreateOrUpdateNannyProfile)

			// Employer Routes
			protected.POST("/profile/employer", profile.CreateOrUpdateEmployerProfile)

			// Shared Routes
			protected.GET("/profile/me", profile.GetMyProfile)
			protected.DELETE("/profile", profile.DeleteProfile)
		}

		//job and booking routes
		// Protected Routes (Requires AuthMiddleware)

		{
			// --- JOB ROUTES ---
			protected.POST("/jobs", job.CreateJob)            // Employer posts
			protected.GET("/jobs/assigned", job.GetNannyJobs) // Nanny sees hers (No Salary)
			protected.POST("/jobs/assign", job.AdminAssign)   // Admin links Nanny to Job

			// --- BOOKING (INTERVIEW) ROUTES ---
			// Nanny Perspective
			protected.POST("/bookings", booking.CreateBooking)    // Idempotent Create
			protected.GET("/bookings/me", booking.GetMyBooking)   // Read
			protected.PUT("/bookings/:id", booking.UpdateBooking) // Update time/date

		}

		{
			// Feedback routes
			protected.POST("/complaints", feedback.FileComplaint)
			protected.POST("/ratings", feedback.RateNanny)
			protected.GET("/ratings/me", feedback.GetMyRatings)

		}

		// Protected routes

		protected.Use(middleware.AuthRequired()) // First, make sure they are logged in
		{
			// --- ADMIN MONITORING (Double Protected) ---
			adminGroup := protected.Group("/admin")
			adminGroup.Use(middleware.AdminOnly()) // Second, make sure they are the ADMIN
			{
				adminGroup.GET("/jobs", job.AdminListJobs)
				adminGroup.GET("/complaints", feedback.AdminListComplaints)
				adminGroup.GET("/ratings", feedback.AdminListAllRatings)
				adminGroup.DELETE("/bookings/:id", booking.AdminDelete)
				adminGroup.POST("/jobs/assign", job.AdminAssign)

				// User Management
				adminGroup.GET("/nannies", admin.ListAllNannies)
				adminGroup.GET("/employers", admin.ListAllEmployers)
				adminGroup.PATCH("/users/status", admin.ToggleUserStatus) // Use PATCH for updates

				// Existing admin routes...
				adminGroup.GET("/complaints", feedback.AdminListComplaints)
			}

			// --- GENERAL USER ROUTES ---
			protected.POST("/complaints", feedback.FileComplaint)
			protected.GET("/jobs/nanny", job.GetNannyJobs)
		}
	}
}
