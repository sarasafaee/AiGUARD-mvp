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


	incomingRoutes.POST("/stream/filter", controller.CreateFilterStream())
	incomingRoutes.GET("/stream/filter/:filter_stream_id", controller.GetFilterStreamByID())
	incomingRoutes.GET("/stream/filters/:stream_id", controller.GetFilterStreamsByStreamID())
	incomingRoutes.PUT("/stream/filter/:filter_stream_id", controller.EditFilterStream())
	incomingRoutes.DELETE("/stream/filter/:filter_stream_id", controller.DeleteFilterStream())



}