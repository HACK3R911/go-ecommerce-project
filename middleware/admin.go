package middleware

import (
	"go-ecommerce-project/database"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userEmail, exists := c.Get("email")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}

		// Проверяем, является ли пользователь админом
		var user struct {
			Is_Admin bool `bson:"is_admin"`
		}
		err := database.UserData(database.Client, "Users").FindOne(
			c,
			bson.M{"email": userEmail},
		).Decode(&user)

		if err != nil || !user.Is_Admin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
			c.Abort()
			return
		}

		c.Next()
	}
}
