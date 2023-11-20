package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/FelipeMarchantVargas/Prueba/controllers"
	"github.com/FelipeMarchantVargas/Prueba/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main(){

	// Configuración de variables de entorno
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("La variable de entorno MONGO_URI no está configurada")
	}
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Error conectando a MongoDB:", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatal("Error cerrando la conexión a MongoDB:", err)
		}
	}()

	uc := controllers.NewUserController(client)

	app := fiber.New()

	origin := os.Getenv("ORIGIN")

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins: origin,
		AllowMethods: "GET,POST,PUT,DELETE",
	}))

	routes.Setup(app, uc)

	// Configuración del puerto
	log.Printf("Iniciando la aplicación en el puerto %s...\n", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Error al iniciar la aplicación:", err)
	}

}