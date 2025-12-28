package main

import (
	"os"
	"qigent/internal/api"
	"qigent/internal/data"

	"github.com/gin-gonic/gin"
)

func main() {
	// Init DB
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/qigent?charset=utf8mb4&parseTime=True&loc=Local"
	}
	if err := data.InitDB(dsn); err != nil {
		// For MVP, maybe allow running without DB if just testing?
		// But we changed everything to use DB. So we must panic or log fatal.
		// log.Fatal("Failed to connect to database: ", err)
		// Let's print but not crash for now if user hasn't set up DB yet?
		// No, it will crash later anyway.
		panic(err)
	}

	// Seed Defaults
	data.SeedRoles()

	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public Routes
	r.POST("/auth/register", api.Register)
	r.POST("/auth/login", api.Login)

	// Protected Routes
	auth := r.Group("/")
	auth.Use(api.AuthMiddleware())
	{
		auth.GET("/ws/chat", api.HandleChat) // Token from query param

		auth.GET("/conversations", api.GetConversations)
		auth.POST("/conversations", api.CreateConversation)
		auth.GET("/conversations/:id", api.GetConversation)
		auth.DELETE("/conversations/:id", api.DeleteConversation)

		auth.GET("/roles", api.GetRoles)
		auth.POST("/roles", api.CreateRole)
		auth.DELETE("/roles/:name", api.DeleteRole)

		auth.GET("/config", api.GetConfig)
		auth.POST("/config", api.UpdateConfig)
	}

	// Read Port from Env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
