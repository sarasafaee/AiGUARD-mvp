package main

import(
	routes "github.com/sarasafaee/AiGUARD-mvp/routes"
	"os"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"path/filepath"
)

func main(){
        dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
        if err != nil {
            log.Fatal(err)
        }
        environmentPath := filepath.Join(dir, ".env")
        err = godotenv.Load(environmentPath)	
	
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
	routes.NotificationRoutes(router)


	// router.GET("/api-1", func(c *gin.Context){
	// 	c.JSON(200, gin.H{"success":"Access granted for api-1"})
	// })


	router.Run(":" + port)
}	
