package routes

import(
	controller "github.com/sarasafaee/sensifai-mvp-crowdsourcing/controllers"
	"github.com/sarasafaee/sensifai-mvp-crowdsourcing/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.Use(middleware.Authenticate())
	
	incomingRoutes.GET("/users", controller.GetUsers())

	incomingRoutes.GET("/user/:user_id", controller.GetUser())
	incomingRoutes.PUT("/user/:user_id", controller.EditUser())

	incomingRoutes.DELETE("/user/deactive", controller.DeactivateUser())

}