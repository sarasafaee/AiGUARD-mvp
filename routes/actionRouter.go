package routes

import(
	controller "github.com/sarasafaee/sensifai-mvp-crowdsourcing/controllers"
	"github.com/sarasafaee/sensifai-mvp-crowdsourcing/middleware"
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