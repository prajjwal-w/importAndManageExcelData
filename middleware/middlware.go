package middleware

import (
	"net/http"
         "log"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
        log.Println(token)
		if token == "" || token != "prajjwal" {
			log.Println("unauthorized")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "provide the auth token"})
			c.Abort()
			return
		}


		c.Next()
	}
}
