package controllers 

import(
"context"
"fmt"
"log"
"strconv"
"net/http"
"time"
"github.com/gin-gonic/gin"
// "github.com/go-playground/validator/v10"
helper "github.com/sarasafaee/sensifai-mvp-crowdsourcing/helpers"
"github.com/sarasafaee/sensifai-mvp-crowdsourcing/models"
"github.com/sarasafaee/sensifai-mvp-crowdsourcing/database"

"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/bson/primitive"
"go.mongodb.org/mongo-driver/mongo"
"go.mongodb.org/mongo-driver/mongo/options"
)

var notificationCollection *mongo.Collection = database.OpenCollection(database.Client, "notification")



//sending notif to one of workerTasks of his own by requester
func SendNotification()gin.HandlerFunc{
	return func(c *gin.Context){
		uid := c.GetString("uid")
		//check access---------------
		if err := helper.CheckUserRole(c, "REQUESTER"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		//----------------------------
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var notification models.Notification

		if err := c.BindJSON(&notification); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var workerTask models.WorkerTask
		err := workerTaskCollection.FindOne(ctx, bson.M{"worker_task_id":notification.Worker_task_id}).Decode(&workerTask)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if workerTask.Last_status == "WAITING"{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this task is not approved yet"})
			return
		}
		var task models.Task
		err = taskCollection.FindOne(ctx, bson.M{"task_id":workerTask.Task_id}).Decode(&task)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if uid != task.Requester_id {
			c.JSON(http.StatusBadRequest, gin.H{"error": " no documents in result"})
			return
		}
		notification.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		notification.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		notification.ID = primitive.NewObjectID()
		notification.Notification_id = notification.ID.Hex()
		notification.Last_status = "SENT"
		validationErr := validate2.Struct(notification)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		}
		//Insert notification into DB---------------------------------
		resultInsertionNumber, insertErr := notificationCollection.InsertOne(ctx, notification)
		if insertErr !=nil {
			msg := fmt.Sprintf("notification item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)
	
	}
}

//geting all the notifs of a workertask by worker
func GetNotifications()gin.HandlerFunc{
	return func(c *gin.Context){
		uid := c.GetString("uid")
		worker_task_id := c.Param("worker_task_id")
		last_status := c.Param("last_status")
		matchStage := bson.D{}

		if last_status == "SENT"{
			matchStage = bson.D{
				{"$match",bson.D{
					{"$and",
						bson.A{
							bson.D{{"worker_task_id",worker_task_id}},
							bson.D{{"last_status", last_status}},


						}},
				},
				},
			}
		}else{
			matchStage = bson.D{{"$match",bson.M{"worker_task_id":worker_task_id}}}
		}
		//check access---------------
		if err := helper.CheckUserRole(c, "WORKER"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		//----------------------------
		//making sure the worker task belongs to worker
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var workerTask models.WorkerTask
		err := workerTaskCollection.FindOne(ctx, bson.M{"worker_task_id":worker_task_id}).Decode(&workerTask)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if workerTask.Worker_id != uid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": " no documents in result"})
			return
		}
		//---------------------------------------
		//returning all messages
		// matchStage := bson.D{{"$match",bson.M{"worker_task_id":worker_task_id}}}

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage <1{
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 !=nil || page<1{
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}}, 
			{"total_count", bson.D{{"$sum", 1}}}, 
			{"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"notif_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},}}}
				result,err := notificationCollection.Aggregate(ctx, mongo.Pipeline{
					matchStage, groupStage, projectStage})
				defer cancel()
				if err!=nil{
					c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while listing user items"})
				}
				var allnotifs []bson.M
				if err = result.All(ctx, &allnotifs); err!=nil{
					log.Fatal(err)
				}
				c.JSON(http.StatusOK, allnotifs[0])
				//change notifs'last status
				upsert := true
				filter := bson.M{"last_status":"SENT"}
				opt := options.UpdateOptions{
					Upsert: &upsert,
				}
				_, err = notificationCollection.UpdateMany(
					ctx,
					filter,
					bson.D{
						{"$set", bson.D{
							{"last_status", "SEEN"},
						}}},
					&opt,
				)
			
				defer cancel()
				if err!=nil{
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
	}
}

// func DeleteNotification()gin.HandlerFunc{
// 	return func(c *gin.Context){

// 	}
// }