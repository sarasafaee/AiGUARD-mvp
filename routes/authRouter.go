package routes

import(
	controller "github.com/sarasafaee/AiGUARD-mvp/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.POST("user/signup", controller.Signup())
	incomingRoutes.POST("user/login", controller.Login())
	incomingRoutes.POST("user/logout", controller.Logout())

	
}