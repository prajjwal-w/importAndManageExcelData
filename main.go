package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prajjwal-w/golang-choicetech/routes"
)

func main() {
	//Creating a gin router
	router := gin.New()

	//Using the default gin logger as middleware
	router.Use(gin.Logger())

	routes.Routes(router)

	//Loading the .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalln("error while loading the .env file")
	}

	port := os.Getenv("PORT")

	//Starting the server
	err := router.Run(":" + port)
	if err != nil {
		log.Fatalf("error while starting the server on port: %v", port)
	}

}
