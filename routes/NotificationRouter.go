package routes

import(
	controller "github.com/sarasafaee/AiGUARD-mvp/controllers"
	"github.com/sarasafaee/AiGUARD-mvp/middleware"
	"github.com/gin-gonic/gin"
)

func NotificationRoutes(incomingRoutes *gin.Engine){
	incomingRoute1 := incomingRoutes.Use(middleware.NotificationSendingAuthenticate())
	incomingRoute1.POST("/notification", controller.SendNotification())
	
	incomingRoutes.Use(middleware.Authenticate())

	// incomingRoutes.GET("/notification/:notification_id", controller.GetNotification())
	incomingRoutes.GET("/notifications/:worker_task_id", controller.GetNotifications())
	incomingRoutes.GET("/notifications/:worker_task_id/:last_status", controller.GetNotifications())

	// incomingRoutes.DELETE("/notification/:notification_id", controller.DeleteNotification())
	incomingRoutes.POST("/notification/token", controller.CreateNotificationToken())
	incomingRoutes.PUT("/notification/approval/:notification_id", controller.NotificationApproval())


}