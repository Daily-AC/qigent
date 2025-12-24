package main

import (
	"qigent/internal/api"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve Frontend (if built) - For MVP we use separate dev servers.
	// But we need CORS if we are on different ports.
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API Routes
	r.GET("/ws/chat", api.HandleChat)
	r.GET("/config", api.GetConfig)
	r.POST("/config", api.UpdateConfig)

	r.Run(":8080")
}
