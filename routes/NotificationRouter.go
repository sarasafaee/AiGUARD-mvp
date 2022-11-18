package routes

import(
	controller "github.com/sarasafaee/sensifai-mvp-crowdsourcing/controllers"
	"github.com/sarasafaee/sensifai-mvp-crowdsourcing/middleware"
	"github.com/gin-gonic/gin"
)

func NotificationRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.Use(middleware.Authenticate())
	
	incomingRoutes.POST("/notification", controller.SendNotification())

	// incomingRoutes.GET("/notification/:notification_id", controller.GetNotification())
	incomingRoutes.GET("/notifications/:worker_task_id", controller.GetNotifications())
	incomingRoutes.GET("/notifications/:worker_task_id/:last_status", controller.GetNotifications())

	// incomingRoutes.DELETE("/notification/:notification_id", controller.DeleteNotification())

}