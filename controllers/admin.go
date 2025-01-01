package controllers

import (
	"context"
	"go-ecommerce-project/models"
	generate "go-ecommerce-project/tokens"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdminController struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewAdminController(prodCollection, userCollection *mongo.Collection) *AdminController {
	return &AdminController{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

func (admin *AdminController) Dashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		productCount, _ := admin.prodCollection.CountDocuments(ctx, bson.M{})
		userCount, _ := admin.userCollection.CountDocuments(ctx, bson.M{})

		// Получаем статистику заказов и доходов
		var revenue float64
		pipeline := mongo.Pipeline{
			{{"$unwind", "$orders"}},
			{{"$group", bson.D{
				{"_id", nil},
				{"total", bson.D{{"$sum", "$orders.Price"}}},
				{"count", bson.D{{"$sum", 1}}},
			}}},
		}

		cursor, err := admin.userCollection.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var result []bson.M
		if err = cursor.All(ctx, &result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var orderCount int64
		if len(result) > 0 {
			revenue = result[0]["total"].(float64)
			orderCount = result[0]["count"].(int64)
		}

		c.HTML(http.StatusOK, "admin.html", gin.H{
			"ProductCount": productCount,
			"UserCount":    userCount,
			"OrderCount":   orderCount,
			"Revenue":      revenue,
		})
	}
}

func (admin *AdminController) Products() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var products []models.Product
		cursor, err := admin.prodCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err = cursor.All(ctx, &products); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "admin-products.html", gin.H{
			"Products": products,
		})
	}
}

func (admin *AdminController) Orders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin.html", gin.H{
			"Content": gin.H{"Orders": []string{}},
		})
	}
}

func (admin *AdminController) Users() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin.html", gin.H{
			"Content": gin.H{"Users": []string{}},
		})
	}
}

func (admin *AdminController) CreateProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var product models.Product
		if err := c.BindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Устанавливаем текущее время в МСК (UTC+3)
		loc, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			loc = time.FixedZone("MSK", 3*60*60) // UTC+3 в секундах
		}
		product.Product_ID = primitive.NewObjectID()
		product.Created_At = time.Now().In(loc)

		_, err = admin.prodCollection.InsertOne(context.Background(), product)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания продукта"})
			return
		}

		c.JSON(http.StatusCreated, product)
	}
}

func (admin *AdminController) UpdateProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")
		objID, _ := primitive.ObjectIDFromHex(productID)

		var product models.Product
		if err := c.BindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := admin.prodCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": objID},
			bson.M{"$set": product},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления продукта"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Продукт обновлен"})
	}
}

func (admin *AdminController) DeleteProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")
		objID, _ := primitive.ObjectIDFromHex(productID)

		_, err := admin.prodCollection.DeleteOne(context.Background(), bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления продукта"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Продукт удален"})
	}
}

func (admin *AdminController) LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем ошибку из query параметров, если она есть
		errorMsg := c.Query("error")
		c.HTML(http.StatusOK, "admin-login.html", gin.H{
			"error": errorMsg,
		})
	}
}

func (admin *AdminController) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		email := c.PostForm("email")
		password := c.PostForm("password")

		var user models.User
		err := admin.userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
		if err != nil {
			c.HTML(http.StatusUnauthorized, "admin-login.html", gin.H{
				"error": "Неверный email или пароль",
			})
			return
		}

		validPassword, _ := VerifyPassword(password, *user.Password)
		if !validPassword {
			c.HTML(http.StatusUnauthorized, "admin-login.html", gin.H{
				"error": "Неверный email или пароль",
			})
			return
		}

		token, refreshToken, _ := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)

		c.SetCookie("admin_token", token, 99999, "/", "", false, true)
		c.SetCookie("refresh_token", refreshToken, 3600, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/admin/dashboard")
	}
}

func (admin *AdminController) NewProductForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin-product-form.html", nil)
	}
}
