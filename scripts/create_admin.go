package main

import (
	"context"
	"go-ecommerce-project/controllers"
	"go-ecommerce-project/database"
	"go-ecommerce-project/models"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	email := "admin2@test.com"
	password := "admin2"
	firstName := "admin2"
	lastName := "admin2"

	hashedPassword := controllers.HashPassword(password)

	admin := models.User{
		ID:         primitive.NewObjectID(),
		Email:      &email,
		Password:   &hashedPassword,
		First_Name: &firstName,
		Last_Name:  &lastName,
		Is_Admin:   true,
	}

	_, err := database.UserData(database.Client, "Users").UpdateOne(
		context.Background(),
		bson.M{"email": email},
		bson.M{"$set": admin},
		options.Update().SetUpsert(true),
	)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Админ успешно создан")
}
