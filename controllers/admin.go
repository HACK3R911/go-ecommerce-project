package controllers

import (
	"context"
	"fmt"
	"go-ecommerce-project/models"
	generate "go-ecommerce-project/tokens"
	"math"
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
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var users []models.User
		cursor, err := admin.userCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err = cursor.All(ctx, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "admin-users.html", gin.H{
			"Users": users,
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
		objID, err := primitive.ObjectIDFromHex(productID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID"})
			return
		}

		var updateData struct {
			Product_Name *string `json:"product_name"`
			Description  *string `json:"description"`
			Manufacturer *string `json:"manufacturer"`
			Price        uint64  `json:"price"`
			Stock        int     `json:"stock"`
			Rating       float64 `json:"rating"`
			Image        *string `json:"image"`
		}

		if err := c.BindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := bson.M{
			"$set": bson.M{
				"product_name": updateData.Product_Name,
				"description":  updateData.Description,
				"manufacturer": updateData.Manufacturer,
				"price":        updateData.Price,
				"stock":        updateData.Stock,
				"rating":       math.Round(updateData.Rating*10) / 10,
				"image":        updateData.Image,
				"updated_at":   time.Now(),
			},
		}

		_, err = admin.prodCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": objID},
			update,
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
		if productID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID продукта не указан"})
			return
		}

		objID, err := primitive.ObjectIDFromHex(productID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID"})
			return
		}

		result, err := admin.prodCollection.DeleteOne(context.Background(), bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления продукта"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Продукт не найден"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Продукт успешно удален"})
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

		c.SetCookie("admin_token", token, 99999, "/", "", false, true) // изменить после теста
		c.SetCookie("refresh_token", refreshToken, 3600, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/admin/dashboard")
	}
}

func (admin *AdminController) NewProductForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin-product-form.html", nil)
	}
}

func (admin *AdminController) DeleteMultipleProducts() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			ProductIds []string `json:"productIds"`
		}

		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
			return
		}

		var objectIds []primitive.ObjectID
		for _, id := range request.ProductIds {
			objID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID"})
				return
			}
			objectIds = append(objectIds, objID)
		}

		result, err := admin.prodCollection.DeleteMany(
			context.Background(),
			bson.M{"_id": bson.M{"$in": objectIds}},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления товаров"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Успешно удалено %d товаров", result.DeletedCount),
		})
	}
}

func (admin *AdminController) DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID пользователя не указан"})
			return
		}

		objID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID"})
			return
		}

		result, err := admin.userCollection.DeleteOne(context.Background(), bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления пользователя"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Пользователь успешно удален"})
	}
}

func (admin *AdminController) DeleteMultipleUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			UserIds []string `json:"userIds"`
		}

		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
			return
		}

		var objectIds []primitive.ObjectID
		for _, id := range request.UserIds {
			objID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID"})
				return
			}
			objectIds = append(objectIds, objID)
		}

		result, err := admin.userCollection.DeleteMany(
			context.Background(),
			bson.M{"_id": bson.M{"$in": objectIds}},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления пользователей"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Успешно удалено %d пользователей", result.DeletedCount),
		})
	}
}

func (admin *AdminController) EditUserForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("id")
		if userId == "" {
			c.HTML(http.StatusBadRequest, "admin-user-edit.html", gin.H{
				"Error": "ID пользователя не указан",
			})
			return
		}

		objID, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			c.HTML(http.StatusBadRequest, "admin-user-edit.html", gin.H{
				"Error": "Неверный формат ID",
			})
			return
		}

		var user models.User
		err = admin.userCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user)
		if err != nil {
			c.HTML(http.StatusNotFound, "admin-user-edit.html", gin.H{
				"Error": "Пользователь не найден",
			})
			return
		}

		c.HTML(http.StatusOK, "admin-user-edit.html", gin.H{
			"User": user,
		})
	}
}

func (admin *AdminController) UpdateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("id")
		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID пользователя не указан"})
			return
		}

		objID, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID"})
			return
		}

		var updateData struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Phone     string `json:"phone"`
			IsAdmin   bool   `json:"is_admin"`
		}

		if err := c.BindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Проверяем, не существует ли уже пользователь с таким email
		var existingUser models.User
		err = admin.userCollection.FindOne(
			context.Background(),
			bson.M{
				"_id":   bson.M{"$ne": objID},
				"email": updateData.Email,
			},
		).Decode(&existingUser)

		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь с таким email уже существует"})
			return
		}

		update := bson.M{
			"$set": bson.M{
				"first_name": updateData.FirstName,
				"last_name":  updateData.LastName,
				"email":      updateData.Email,
				"phone":      updateData.Phone,
				"is_admin":   updateData.IsAdmin,
				"updated_at": time.Now(),
			},
		}

		result, err := admin.userCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": objID},
			update,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении пользователя"})
			return
		}

		if result.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Пользователь успешно обновлен"})
	}
}

func (admin *AdminController) EditProductForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")
		if productID == "" {
			c.HTML(http.StatusBadRequest, "admin-product-edit.html", gin.H{
				"Error": "ID товара не указан",
			})
			return
		}

		objID, err := primitive.ObjectIDFromHex(productID)
		if err != nil {
			c.HTML(http.StatusBadRequest, "admin-product-edit.html", gin.H{
				"Error": "Неверный формат ID",
			})
			return
		}

		var product models.Product
		err = admin.prodCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&product)
		if err != nil {
			c.HTML(http.StatusNotFound, "admin-product-edit.html", gin.H{
				"Error": "Товар не найден",
			})
			return
		}

		c.HTML(http.StatusOK, "admin-product-edit.html", gin.H{
			"Product": product,
		})
	}
}
