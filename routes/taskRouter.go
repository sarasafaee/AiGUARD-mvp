package routes

import(
	controller "github.com/sarasafaee/AiGUARD-mvp/controllers"
	"github.com/sarasafaee/AiGUARD-mvp/middleware"
	"github.com/gin-gonic/gin"
)

func TaskRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.Use(middleware.Authenticate())
	
	incomingRoutes.POST("/task", controller.CtreateTask())
	incomingRoutes.GET("/task/:task_id", controller.GetTask())
	incomingRoutes.PUT("/task/:task_id", controller.EditTask())
	incomingRoutes.DELETE("/task/:task_id", controller.DeleteTask())

	incomingRoutes.GET("/tasks", controller.GetTasks())
	incomingRoutes.GET("/tasks/customized", controller.GetCustomizedTasks())


	incomingRoutes.POST("/task/filter", controller.CreateFilterTask())
	incomingRoutes.GET("/task/filter/:filter_task_id", controller.GetFilterTaskByID())
	incomingRoutes.GET("/task/filtertask/:task_id", controller.GetFilterTaskByTaskID())
	incomingRoutes.PUT("/task/filter/:filter_task_id", controller.EditFilterTask())
	incomingRoutes.DELETE("/task/filter/:filter_task_id", controller.DeleteFilterTask())

	incomingRoutes.POST("/task/application/:task_id", controller.ApplyTask())
	incomingRoutes.PUT("/task/evaluation/:worker_task_id", controller.EvaluateWorkerTask())
	incomingRoutes.GET("/task/workertasks/:last_status", controller.GetWorkerTasks())
	incomingRoutes.GET("/task/workertasks", controller.GetWorkerTasks())








}