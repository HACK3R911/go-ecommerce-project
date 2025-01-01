package middleware

import (
	"net/http"

	token "go-ecommerce-project/tokens"

	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем сначала куки
		ClientToken, err := c.Cookie("admin_token")
		if err != nil {
			// Если куки нет, проверяем заголовок
			ClientToken = c.Request.Header.Get("token")
			if ClientToken == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
				c.Abort()
				return
			}
		}

		claims, msg := token.ValidateToken(ClientToken)
		if msg != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("uid", claims.Uid)
		c.Next()
	}
}
