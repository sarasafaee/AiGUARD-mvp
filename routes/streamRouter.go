package routes

import(
	controller "github.com/sarasafaee/AiGUARD-mvp/controllers"
	"github.com/sarasafaee/AiGUARD-mvp/middleware"
	"github.com/gin-gonic/gin"
)

func StreamRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.Use(middleware.Authenticate())
	
	incomingRoutes.POST("/stream", controller.CtreateStream())
	incomingRoutes.GET("/stream/:stream_id", controller.GetStream())
	incomingRoutes.PUT("/stream/:stream_id", controller.EditStream())
	incomingRoutes.DELETE("/stream/:stream_id", controller.DeleteStream())

	incomingRoutes.GET("/streams", controller.GetStreams())


	incomingRoutes.POST("/stream/activity", controller.CreateActivityStream())
	incomingRoutes.GET("/stream/activity/:activity_stream_id", controller.GetActivityStreamByID())
	incomingRoutes.GET("/stream/activities/:stream_id", controller.GetActivityStreamsByStreamID())
	incomingRoutes.PUT("/stream/activity/:activity_stream_id", controller.EditActivityStream())
	incomingRoutes.DELETE("/stream/activity/:activity_stream_id", controller.DeleteActivityStream())



}