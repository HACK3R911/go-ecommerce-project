package main

import (
	"log"
	"net/http"
	"os"

	"go-ecommerce-project/controllers"
	"go-ecommerce-project/database"
	"go-ecommerce-project/middleware"
	"go-ecommerce-project/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	app := controllers.NewApplication(database.ProductData(database.Client, "Products"), database.UserData(database.Client, "Users"))
	adminApp := controllers.NewAdminController(database.ProductData(database.Client, "Products"), database.UserData(database.Client, "Users"))

	router := gin.New()
	router.Use(gin.Logger())

	// Загружаем HTML шаблоны
	router.LoadHTMLGlob("templates/*")

	// Публичные маршруты (без аутентификации)
	router.GET("/", controllers.HomePage())

	// Маршруты пользователя
	routes.UserRoutes(router)

	// Защищенные маршруты (требуют аутентификации)
	protected := router.Group("/")
	protected.Use(middleware.Authentication())
	{
		protected.GET("/addtocart", app.AddToCart())
		protected.GET("/removeitem", app.RemoveItem())
		protected.GET("/listcart", controllers.GetItemFromCart())
		protected.POST("/addaddress", controllers.AddAddress())
		protected.PUT("/edithomeaddress", controllers.EditHomeAddress())
		protected.PUT("/editworkaddress", controllers.EditWorkAddress())
		protected.GET("/deleteaddresses", controllers.DeleteAddress())
		protected.GET("/cartcheckout", app.BuyFromCart())
		protected.GET("/instantbuy", app.InstantBuy())
	}

	router.GET("/admin/login", adminApp.LoginPage())
	router.POST("/admin/login", adminApp.Login())

	admin := router.Group("/admin")
	admin.Use(middleware.Authentication(), middleware.AdminAuth())
	{
		admin.GET("/dashboard", adminApp.Dashboard())
		admin.GET("/products", adminApp.Products())
		admin.GET("/orders", adminApp.Orders())
		admin.GET("/users", adminApp.Users())

		admin.GET("/products/new", adminApp.NewProductForm())
		admin.POST("/products", adminApp.CreateProduct())
		admin.PUT("/products/:id", adminApp.UpdateProduct())
		admin.GET("/products/:id/edit", adminApp.EditProductForm())
		admin.DELETE("/products/:id", adminApp.DeleteProduct())
		admin.POST("/products/batch-delete", adminApp.DeleteMultipleProducts())

		admin.DELETE("/users/:id", adminApp.DeleteUser())
		admin.GET("/users/:id/edit", adminApp.EditUserForm())
		admin.PUT("/users/:id", adminApp.UpdateUser())
		admin.DELETE("/users/bulk-delete", adminApp.DeleteMultipleUsers())
	}

	router.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", nil)
	})

	log.Fatal(router.Run(":" + port))
}
