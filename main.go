package main

import(
	routes "github.com/sarasafaee/sensifai-mvp-crowdsourcing/routes"
	"os"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main(){
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")

	if port==""{
		port="8000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	routes.TaskRoutes(router)
	routes.StreamRoutes(router)
	routes.ActionRoutes(router)



	router.GET("/api-1", func(c *gin.Context){
		c.JSON(200, gin.H{"success":"Access granted for api-1"})
	})

	router.GET("/api-2", func(c *gin.Context){
		c.JSON(200, gin.H{"success":"Access granted for api-2"})
	})

	router.Run(":" + port)
}	