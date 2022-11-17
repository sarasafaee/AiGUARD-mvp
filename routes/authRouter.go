package routes

import(
	controller "github.com/sarasafaee/sensifai-mvp-crowdsourcing/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.POST("users/signup", controller.Signup())
	incomingRoutes.POST("users/login", controller.Login())
	incomingRoutes.POST("users/logout", controller.Logout())

	
}