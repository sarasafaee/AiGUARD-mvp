package routes

import(
	controller "github.com/sarasafaee/AiGUARD-mvp/controllers"
	"github.com/sarasafaee/AiGUARD-mvp/middleware"
	"github.com/gin-gonic/gin"
)

func ActionRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.Use(middleware.Authenticate())
	
	incomingRoutes.POST("/action", controller.CtreateAction())
	incomingRoutes.GET("/action/:action_id", controller.GetAction())
	incomingRoutes.PUT("/action/:action_id", controller.EditAction())
	incomingRoutes.DELETE("/action/:action_id", controller.DeleteAction())

	incomingRoutes.GET("/actions", controller.GetActions())


}