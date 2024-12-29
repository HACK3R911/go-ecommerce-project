package main

import (
	"log"
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

	log.Fatal(router.Run(":" + port))
}
