package controllers

import (
	"context"
	"net/http"
	"time"

	"go-ecommerce-project/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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

		// Получаем популярные товары
		var featuredProducts []models.Product
		featuredCursor, err := ProductCollection.Find(ctx, bson.M{"rating": bson.M{"$gte": 4}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении популярных товаров"})
			return
		}
		if err = featuredCursor.All(ctx, &featuredProducts); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обработке популярных товаров"})
			return
		}

		// Получаем новые поступления
		var newArrivals []models.Product
		newArrivalsCursor, err := ProductCollection.Find(ctx, bson.M{},
			options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(8))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении новых товаров"})
			return
		}
		if err = newArrivalsCursor.All(ctx, &newArrivals); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обработке новых товаров"})
			return
		}

		// Заглушка для категорий и баннеров
		categories := []string{"Электроника", "Одежда", "Книги", "Спорт", "Дом и сад"}
		bannerImages := []string{
			"/banners/banner1.jpg",
			"/banners/banner2.jpg",
			"/banners/banner3.jpg",
		}

		response := HomePageResponse{
			FeaturedProducts: featuredProducts,
			NewArrivals:      newArrivals,
			Categories:       categories,
			BannerImages:     bannerImages,
		}

		c.HTML(http.StatusOK, "home.html", response)
	}
}
