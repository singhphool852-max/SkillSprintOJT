package main

import (
	"log"
	"os"
 update-cors-amplify

	"strings"
	"time"
 main

	"backend/arena"
	"backend/database"
	"backend/handlers"
	"backend/leaderboard"
	"backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("[ENV] No .env file found, using system environment variables")
	} else {
		log.Println("[ENV] Successfully loaded configuration from .env")
	}

	database.ConnectDB()

	// Start leaderboard WebSocket hub
	hub := leaderboard.NewHub()
	handlers.LeaderboardHub = hub

	// Start arena session WebSocket hub
	arenaHub := arena.NewSessionHub()
	handlers.ArenaSessionHub = arenaHub

	// Start background auto-submit watcher for expired attempts
	handlers.StartAutoSubmitWatcher()

	r := gin.Default()

	// Setup Robust CORS to allow Next.js communication
	config := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Allow localhost, Vercel, and Amplify domains
			return origin == "http://localhost:3000" || 
				   strings.HasSuffix(origin, ".vercel.app") || 
				   strings.HasSuffix(origin, ".amplifyapp.com") ||
				   origin == "https://skillsprintojt.onrender.com"
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(config))

	// Explicitly handle OPTIONS for all paths
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(200)
	})

	// Health and Root Endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "Neural Link Stable", "time": time.Now().Format(time.RFC3339)})
	})
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "SkillSprint Neural Engine Online",
			"version": "1.0.0",
			"api_root": "/api",
		})
	})

	api := r.Group("/api")
	
	// Administrative Utility
	api.GET("/admin/reload-env", func(c *gin.Context) {
		godotenv.Load()
		c.JSON(200, gin.H{"status": "Neural Configuration Refreshed"})
	})

	// Auth Routes
	auth := api.Group("/auth")
	{
		auth.POST("/login", handlers.LoginHandler)
		auth.POST("/signup", handlers.SignupHandler)
		auth.POST("/google", handlers.GoogleLoginHandler)
		auth.POST("/google/login", handlers.GoogleLoginHandler) // Alias for user suggestion
		auth.POST("/logout", handlers.LogoutHandler)

		// Protected me route
		auth.GET("/me", middleware.JWTMiddleware(), handlers.MeHandler)
	}

	// Root-level aliases (Failsafe for mismatched frontend paths)
	r.POST("/google", handlers.GoogleLoginHandler)
	r.POST("/login", handlers.LoginHandler)
	r.GET("/me", middleware.JWTMiddleware(), handlers.MeHandler)

	// Public Arena Routes
	api.GET("/arenas", handlers.GetArenas)
	api.GET("/arenas/:id", handlers.GetArenaDetail)
	api.GET("/arenas/:id/quizzes", handlers.GetArenaQuizzes)
	api.GET("/quizzes/:quizId/questions", handlers.GetQuizQuestions)

	// Leaderboard Routes
	api.GET("/attempts/leaderboard", handlers.GetLeaderboard)
	api.GET("/leaderboard/global", handlers.GetGlobalLeaderboard)

	// Public Topics (for Arena)
	api.GET("/topics", handlers.ListPublicTopics)
	api.GET("/topics/:slug/tests", handlers.ListPublicTestsByTopic)

	// Training Session (serves real DB-backed questions)
	api.POST("/train/session", handlers.CreateTrainSession)
	api.GET("/train/session/:id", handlers.GetTrainSessionDetail)

	// Protected Routes (Attempts & Evaluation)
	protected := api.Group("/")
	protected.Use(middleware.JWTMiddleware())
	{
		protected.POST("/attempts", handlers.SubmitAttempt)
		protected.GET("/attempts/:id", handlers.GetAttemptResult)
		protected.POST("/evaluate-answer", handlers.EvaluateAnswer)
		protected.POST("/training/verify", handlers.VerifyAnswer)
		protected.POST("/training/generate", handlers.GenerateTrainingSession)
		// User dashboard & results
		protected.GET("/dashboard/stats", handlers.GetUserDashboardStats)
		protected.GET("/dashboard/full", handlers.GetUserDashboardFull)
		protected.GET("/results", handlers.ListUserResults)
		protected.GET("/results/:attemptId", handlers.GetTestResult)

		// Training module additions
		protected.POST("/train/upload-notes", handlers.UploadNotes)
		protected.GET("/training/session/:id", handlers.GetTrainingSession)
	}

	// Admin Routes (JWT + Admin role required)
	admin := api.Group("/admin")
	admin.Use(middleware.JWTMiddleware())
	admin.Use(middleware.AdminOnly())
	{
		// Topic CRUD
		admin.POST("/topics", handlers.CreateTopic)
		admin.GET("/topics", handlers.ListTopics)
		admin.PUT("/topics/:id", handlers.UpdateTopic)
		admin.DELETE("/topics/:id", handlers.DeleteTopic)

		// Test CRUD
		admin.POST("/tests", handlers.CreateTest)
		admin.GET("/tests", handlers.ListTests)
		admin.GET("/tests/:id", handlers.GetTestDetail)
		admin.PUT("/tests/:id", handlers.UpdateTest)
		admin.DELETE("/tests/:id", handlers.DeleteTest)
		admin.PATCH("/tests/:id/publish", handlers.PublishTest)
		admin.PATCH("/tests/:id/activate", handlers.ActivateTest)

		// Question management
		admin.POST("/tests/:id/questions", handlers.AddQuestion)
		admin.GET("/tests/:id/questions", handlers.ListQuestions)
		admin.PUT("/questions/:id", handlers.UpdateQuestion)
		admin.DELETE("/questions/:id", handlers.DeleteQuestion)

		// Testcase management
		admin.POST("/questions/:id/testcases", handlers.AddTestcase)
		admin.DELETE("/testcases/:id", handlers.DeleteTestcase)

		// Dashboard analytics
		admin.GET("/dashboard/stats", handlers.GetAdminDashboardStats)
		admin.GET("/analytics", handlers.GetAdminDashboardStats)
		admin.GET("/dashboard/recent", handlers.GetRecentActivity)
	}


	// Arena Test Routes (JWT required)
	arena := api.Group("/arena")
	arena.Use(middleware.JWTMiddleware())
	{
		arena.GET("/active", handlers.GetActiveTest)
		arena.GET("/tests", handlers.ListPublishedTests)
		arena.GET("/languages", handlers.GetLanguages)
		arena.POST("/tests/:id/join", handlers.JoinTest)
		arena.GET("/attempts/:id", handlers.GetTestAttempt)
		arena.POST("/submissions/mcq", handlers.SaveMCQ)
		arena.POST("/submissions/run", handlers.RunCode)
		arena.POST("/submissions/code", handlers.SubmitCode)
		arena.POST("/submissions/draft", handlers.SaveDraft)
		arena.POST("/attempts/:id/submit", handlers.SubmitTestAttempt)
		arena.GET("/attempts/:id/status", handlers.GetAttemptStatus)
		arena.POST("/violations", handlers.LogViolation)
	}

	// Leaderboard Routes
	api.GET("/leaderboard/:testId", handlers.GetTestLeaderboard)

	// WebSocket route (outside /api — no JSON middleware needed)
	r.GET("/ws/leaderboard/:testId", handlers.LeaderboardWS)
	r.GET("/ws/arena/:attemptId", handlers.ArenaSessionWS)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Gin server on 0.0.0.0:%s", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
