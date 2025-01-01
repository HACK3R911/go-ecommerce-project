package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"go-ecommerce-project/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HomePageResponse struct {
	FeaturedProducts []models.Product `json:"featured_products"`
	NewArrivals      []models.Product `json:"new_arrivals"`
	Categories       []string         `json:"categories"`
	BannerImages     []string         `json:"banner_images"`
}

func HomePage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Получаем информацию о пользователе из сессии
		userEmail, exists := c.Get("email")
		var userData *models.User
		var cartItemsCount int64 = 0

		if exists {
			var err error
			userData, err = getUserByEmail(ctx, userEmail.(string))
			if err != nil {
				log.Printf("Error getting user data: %v", err)
			}

			// Преобразуем строковый ID в ObjectID
			userObjectID, err := primitive.ObjectIDFromHex(userData.User_ID)
			if err != nil {
				log.Printf("Error converting user ID: %v", err)
			} else {
				cartItemsCount, err = getCartItemsCount(ctx, userObjectID)
				if err != nil {
					log.Printf("Error getting cart items count: %v", err)
				}
			}
		}

		// Получаем популярные товары и новинки
		featuredProducts, err := getFeaturedProducts(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении популярных товаров"})
			return
		}

		newArrivals, err := getNewArrivals(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении новинок"})
			return
		}

		c.HTML(http.StatusOK, "home.html", gin.H{
			"User":             userData,
			"CartItemsCount":   cartItemsCount,
			"FeaturedProducts": featuredProducts,
			"NewArrivals":      newArrivals,
		})
	}
}

func getUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func getCartItemsCount(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	count, err := ProductCollection.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func getFeaturedProducts(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	cursor, err := ProductCollection.Find(ctx, bson.M{"rating": bson.M{"$gte": 4}})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func getNewArrivals(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(10)
	cursor, err := ProductCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}
