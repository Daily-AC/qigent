package api

import (
	"qigent/internal/data"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var SecretKey = []byte("my_secret_key") // Move to config in prod

func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	user, err := data.CreateUser(req.Username, string(hash))
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(200, user)
}

func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	user, err := data.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Invalid password"})
		return
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "qigent",
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, _ := token.SignedString(SecretKey)

	c.JSON(200, gin.H{"token": tokenString, "username": user.Username})
}

// AuthMiddleware validates the token and sets UserID in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Check Query Param (for WebSocket)
			authHeader = c.Query("token")
		} else {
			authHeader = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthenticated"})
			return
		}

		token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
			return SecretKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		userID := uint(claims["sub"].(float64))
		c.Set("userID", userID)
		c.Next()
	}
}
