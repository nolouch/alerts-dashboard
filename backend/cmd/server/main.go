package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nolouch/alerts-platform-v2/internal/api"
	"github.com/nolouch/alerts-platform-v2/internal/db"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found or unable to load .env file")
	} else {
		log.Println("✅ Loaded environment variables from .env file")
	}

	// Initialize Database
	if err := db.Init(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	r := gin.Default()

	// CORS Configuration (Allow Frontend)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all for dev simplicity (ports change)
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API Routes
	v1 := r.Group("/api")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Components Endpoints
		v1.GET("/categories", api.GetCategories)
		v1.GET("/components", api.GetComponents)
		v1.GET("/components/:name/stats", api.GetComponentStats)
		v1.GET("/components/:name/rules", api.GetComponentRules)
		v1.PUT("/components/:name/rules", api.UpdateComponentRule)

		// New Dashboard Route
		v1.GET("/dashboard", api.GetDashboardData)
		v1.GET("/dashboard/issues", api.GetDashboardIssues)
		v1.POST("/issues/:id/mute", api.MuteIssue)
		// New Rules Notify Manager Routes
		v1.GET("/rules-notify-manager", api.GetRulesNotifyConfig)
		v1.PUT("/rules-notify-manager", api.UpdateRulesNotifyConfig)

		// Rule Tasks Routes

		v1.GET("/tasks", api.HandleGetTasks)
		v1.POST("/tasks", api.HandleCreateTask)
	}

	// Serve Frontend Static Files (for production/release)
	// Only serves if "public" directory exists (created by release process)
	if _, err := os.Stat("./public"); err == nil {
		log.Println("Detected 'public' directory, serving static files")
		r.Static("/assets", "./public/assets")

		// Serve other root files if needed, or rely on NoRoute for SPA fallthrough
		// Handling SPA client-side routing
		r.NoRoute(func(c *gin.Context) {
			// If it's an API 404, return JSON
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(404, gin.H{"error": "API endpoint not found"})
				return
			}
			// Otherwise serve index.html
			c.File("./public/index.html")
		})
	}

	// Register Update Routes (for JIRA data sync)
	// Register Update Routes (for JIRA data sync)
	api.RegisterUpdateRoutes(r, db.DB)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8818"
	}
	host := os.Getenv("HOST")
	addr := host + ":" + port

	log.Printf("Server running on %s", addr)
	r.Run(addr)
}
