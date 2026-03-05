package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kolesov-ai/MagicStreamMovies/Server/MagicStreamMoviesServer/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := utils.GetAccessToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is empty", "detail": err.Error()})
			c.Abort()
			return
		}
		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("userId", claims.UserId)
		c.Set("role", claims.Role)
		c.Next()
	}
}
