package controllers 

import(
"context"
"fmt"
"log"
"strconv"
"net/http"
"time"
"github.com/gin-gonic/gin"
"github.com/go-playground/validator/v10"
helper "github.com/sarasafaee/sensifai-mvp-crowdsourcing/helpers"
"github.com/sarasafaee/sensifai-mvp-crowdsourcing/models"
"github.com/sarasafaee/sensifai-mvp-crowdsourcing/database"

"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/bson/primitive"
"go.mongodb.org/mongo-driver/mongo"
"go.mongodb.org/mongo-driver/mongo/options"
)

var taskCollection *mongo.Collection = database.OpenCollection(database.Client, "task")
var filterTaskCollection *mongo.Collection = database.OpenCollection(database.Client, "filterTask")
var workerTaskCollection *mongo.Collection = database.OpenCollection(database.Client, "workerTask")
var taskStatusCollection *mongo.Collection = database.OpenCollection(database.Client, "taskStatus")


var validate1 = validator.New()

//create a task by requester
func CtreateTask()gin.HandlerFunc{

	return func(c *gin.Context){
		uid := c.GetString("uid")
		//check access---------------
		if err := helper.CheckUserRole(c, "REQUESTER"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		//----------------------------
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var task models.Task

		if err := c.BindJSON(&task); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//custom validation----------
		today, _  := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if today.Before(task.Deadline) == false {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"deadline can not be in the past"})
			return
		}
		//---------------------------
		validationErr := validate1.Struct(task)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
		} else {
			//stream Existence in DB -----------------------
			streamCount, err := streamCollection.CountDocuments(ctx, bson.M{"stream_id":task.Stream_id})
			defer cancel()
			if err != nil {
				log.Panic(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while checking for the stream"})
				return
			}
			if streamCount ==0{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"this stream is not defined"})
				return
			}
			//stream existence in tasks-----------------------
			streamTaskCount, err := taskCollection.CountDocuments(ctx, bson.M{"stream_id":task.Stream_id})
			defer cancel()
			if err != nil {
				log.Panic(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while checking for the stream"})
				return
			}

			if streamTaskCount >0{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"task for this stream already exists"})
			}else{
				task.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				task.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				task.ID = primitive.NewObjectID()
				task.Task_id = task.ID.Hex()
				task.Requester_id = uid
				//Insert task into DB---------------------------------
				resultInsertionNumber, insertErr := taskCollection.InsertOne(ctx, task)
				if insertErr !=nil {
					msg := fmt.Sprintf("Task item was not created")
					c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
					return
				}
				defer cancel()
				c.JSON(http.StatusOK, resultInsertionNumber)
			}
	
		}

	}

}

//get a task just for requesterowner and admin
func GetTask()gin.HandlerFunc{

	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		taskId := c.Param("task_id")
		var task models.Task
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		if UserRole == "WORKER" {
			c.JSON(http.StatusBadRequest, gin.H{"error" : "Unauthorized to access this resource"})
			return
		}else if UserRole == "REQUESTER" {
			//check task existence
			err := taskCollection.FindOne(ctx, bson.M{"task_id":taskId}).Decode(&task)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			//check task belonging to owner
			if task.Requester_id == uid {
				c.JSON(http.StatusOK, task)
				return
			}else{
				c.JSON(http.StatusInternalServerError, gin.H{"error": " no documents in result"})
				return
			}
		}else if UserRole == "ADMIN"{
			//check task existence
			err := taskCollection.FindOne(ctx, bson.M{"task_id":taskId}).Decode(&task)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, task)

		}
		

	}
}

//get all tasks just for requesterowner and admin
func GetTasks()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		matchStage := bson.D{}
		if UserRole == "ADMIN" {
			matchStage = bson.D{{"$match",bson.M{}}}

		}else if UserRole == "REQUESTER" {
			matchStage = bson.D{{"$match",bson.M{"requester_id":uid}}}
		}else if UserRole == "WORKER" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized to access this resource"})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		
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
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},}}}
				result,err := taskCollection.Aggregate(ctx, mongo.Pipeline{
					matchStage, groupStage, projectStage})
				defer cancel()
				if err!=nil{
					c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while listing user items"})
				}
				var allusers []bson.M
				if err = result.All(ctx, &allusers); err!=nil{
					log.Fatal(err)
				}
				c.JSON(http.StatusOK, allusers[0])
	}
}

//get all GetCustomizedTasks for worker
func GetCustomizedTasks()gin.HandlerFunc{
	return func(c *gin.Context){

		uid := c.GetString("uid")
		var user models.User
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		if err := helper.CheckUserRole(c, "WORKER"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		err := userCollection.FindOne(ctx, bson.M{"user_id":uid}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
			return
		}

	    matchStage := bson.D{
			{"$lookup",
			bson.M{
			   "from": "filterTask",
			   "localField": "task_id",
			   "foreignField": "task_id",
			   "as": "filter",
			}},

		}
		matchStage2 := bson.D{
			{"$match",bson.D{
				{"$or",
					bson.A{
						bson.D{{"filter.tags", user.Location}},
						bson.D{{"filter", bson.D{
							{"$size", 0},}}},
					}},
			},
			},
		}
		  
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage <1{
			recordPerPage = 10000
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 !=nil || page<1{
			page = 1
		}
		startIndex := (page - 1) * recordPerPage
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}}, 
			{"total_count", bson.D{{"$sum", 1}}}, 
			{"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}
		result,err := taskCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage,matchStage2, groupStage, projectStage})

		defer cancel()
		if err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while listing user items"})
		}
		var allusers []bson.M
		if err = result.All(ctx, &allusers); err!=nil{
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allusers[0])
		}
}

//edit a task by requesterowner and admin
func EditTask()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")

		var task models.Task
		if UserRole == "WORKER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized to access this resource"})
			return
		}
		taskId := c.Param("task_id")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		err := taskCollection.FindOne(ctx, bson.M{"task_id":taskId}).Decode(&task)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if task.Requester_id != uid && UserRole == "REQUESTER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error":" no documents in result"})
			return
		}
		if err := c.BindJSON(&task); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//custom validation----------
		today, _  := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if today.Before(task.Deadline) == false {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"deadline can not be in the past"})
			return
		}
		// --------------------------
		task.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		validationErr := validate1.Struct(task)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		}
		upsert := true
		filter := bson.M{"task_id":taskId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
	
		_, err = taskCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", task},
			},
			&opt,
		)
	
		defer cancel()
	
		if err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, task)


	}
}

//delete a task and all its dependencies(task_filter) by requesterowner and admin
func DeleteTask()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")

		var task models.Task
		if UserRole == "WORKER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized to access this resource"})
			return
		}
		taskId := c.Param("task_id")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		err := taskCollection.FindOne(ctx, bson.M{"task_id":taskId}).Decode(&task)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if UserRole == "REQUESTER"{
			if task.Requester_id != uid {
				c.JSON(http.StatusInternalServerError, gin.H{"error":" no documents in result"})
				return
			}	
		}

		_, err = taskCollection.DeleteOne(ctx, bson.M{"task_id":taskId})
		_, err = filterTaskCollection.DeleteOne(ctx, bson.M{"task_id":taskId})//might cause BUG

		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while deleting task"})
		}
		c.JSON(http.StatusOK,gin.H{"message":"task deleted successfully"} )
	}
}

//create a filterTask by requester
func CreateFilterTask()gin.HandlerFunc{
	return func(c *gin.Context){
		uid := c.GetString("uid")
		//check access---------------
		if err := helper.CheckUserRole(c, "REQUESTER"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		//----------------------------
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var filterTask models.FilterTask

		if err := c.BindJSON(&filterTask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate1.Struct(filterTask)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		} else {
			//task Existence in DB -----------------------

			var task models.Task
			err := taskCollection.FindOne(ctx, bson.M{"task_id":filterTask.Task_id}).Decode(&task)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			//check task belonging to owner
			if task.Requester_id != uid {
				c.JSON(http.StatusInternalServerError, gin.H{"error": " no documents in result"})
				return
			}

			//filterTask existence in DB-----------------------
			filterTaskCount, err := filterTaskCollection.CountDocuments(ctx, bson.M{"task_id":filterTask.Task_id})
			defer cancel()
			if err != nil {
				log.Panic(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while checking for the stream"})
				return
			}

			if filterTaskCount >0{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"filterTask for this task already exists"})
			}else{
				filterTask.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				filterTask.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				filterTask.ID = primitive.NewObjectID()
				filterTask.Filter_task_id = filterTask.ID.Hex()
				//Insert task into DB---------------------------------
				resultInsertionNumber, insertErr := filterTaskCollection.InsertOne(ctx, filterTask)
				if insertErr !=nil {
					msg := fmt.Sprintf("filterTask item was not created")
					c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
					return
				}
				defer cancel()
				c.JSON(http.StatusOK, resultInsertionNumber)
			}
	
		}

	}
}

//get filterTask By filterTaskId
func GetFilterTaskByID()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		filterTaskId := c.Param("filter_task_id")
		var filterTask models.FilterTask
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		//check filterTask existence
		err := filterTaskCollection.FindOne(ctx, bson.M{"filter_task_id":filterTaskId}).Decode(&filterTask)
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if UserRole == "WORKER" {
			c.JSON(http.StatusBadRequest, gin.H{"error" : "Unauthorized to access this resource"})
			return
		}else if UserRole == "REQUESTER" {
			//check filterTask and task belonging to owner
			var task models.Task
			err := taskCollection.FindOne(ctx, bson.M{"task_id":filterTask.Task_id}).Decode(&task)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if task.Requester_id == uid {
				c.JSON(http.StatusOK, filterTask)
				return
			}else{
				c.JSON(http.StatusInternalServerError, gin.H{"error": " no documents in result"})
				return
			}
		}else if UserRole == "ADMIN"{
			c.JSON(http.StatusOK, filterTask)

		}
		
	}
}

//get filterTask By taskId
func GetFilterTaskByTaskID()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		taskId := c.Param("task_id")
		var filterTask models.FilterTask
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		//check filterTask existence
		err := filterTaskCollection.FindOne(ctx, bson.M{"task_id":taskId}).Decode(&filterTask)
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if UserRole == "WORKER" {
			c.JSON(http.StatusBadRequest, gin.H{"error" : "Unauthorized to access this resource"})
			return
		}else if UserRole == "REQUESTER" {
			//check filterTask and task belonging to owner
			var task models.Task
			err := taskCollection.FindOne(ctx, bson.M{"task_id":filterTask.Task_id}).Decode(&task)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if task.Requester_id == uid {
				c.JSON(http.StatusOK, filterTask)
				return
			}else{
				c.JSON(http.StatusInternalServerError, gin.H{"error": " no documents in result"})
				return
			}
		}else if UserRole == "ADMIN"{
			c.JSON(http.StatusOK, filterTask)

		}
		
	}
}

//edit a filtertask by requesterowner and admin
func EditFilterTask()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")

		var filterTask models.FilterTask
		if UserRole == "WORKER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized to access this resource"})
			return
		}
		filterTaskId := c.Param("filter_task_id")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		err := filterTaskCollection.FindOne(ctx, bson.M{"filter_task_id":filterTaskId}).Decode(&filterTask)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var task models.Task

		if UserRole == "REQUESTER"{
			err := taskCollection.FindOne(ctx, bson.M{"task_id":filterTask.Task_id}).Decode(&task)
			defer cancel()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if task.Requester_id != uid {
				c.JSON(http.StatusInternalServerError, gin.H{"error":" no documents in result"})
				return
			}

			
		}
		if err := c.BindJSON(&filterTask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//update updated_at --------------------------
		filterTask.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		validationErr := validate1.Struct(filterTask)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		}
		//costum validation 
		if filterTask.Task_id != task.Task_id || filterTaskId != filterTask.Filter_task_id{
			c.JSON(http.StatusBadRequest, gin.H{"error":"editing TaskId and filterTaskId is not possible"})
			return
		} 
		upsert := true
		filter := bson.M{"filter_task_id":filterTaskId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		// ---------------------------------------
		_, err = filterTaskCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", filterTask},
			},
			&opt,
		)
	
		defer cancel()
		if err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, filterTask)


	}
}

//delete a filterTask by requesterowner and admin
func DeleteFilterTask()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")

		var filterTask models.FilterTask
		if UserRole == "WORKER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized to access this resource"})
			return
		}
		filterTaskId := c.Param("filter_task_id")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		err := filterTaskCollection.FindOne(ctx, bson.M{"filter_task_id":filterTaskId}).Decode(&filterTask)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var task models.Task

		if UserRole == "REQUESTER"{
			err := taskCollection.FindOne(ctx, bson.M{"task_id":filterTask.Task_id}).Decode(&task)
			defer cancel()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if task.Requester_id != uid {
				c.JSON(http.StatusInternalServerError, gin.H{"error":" no documents in result"})
				return
			}	
		}
		_, err = filterTaskCollection.DeleteOne(ctx, bson.M{"filter_task_id":filterTaskId})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while deleting filter"})
		}
		c.JSON(http.StatusOK,gin.H{"message":"filter deleted successfully"} )


	}
}

//apply to contribute to a task by worker
//worker can apply several times for a task.(good or bad?)
func ApplyTask()gin.HandlerFunc{
	return func(c *gin.Context){
		uid := c.GetString("uid")
		taskId := c.Param("task_id")

		var workerTask models.WorkerTask
		var taskStatus models.TaskStatus
		var task models.Task

		if err := helper.CheckUserRole(c, "WORKER"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		err := taskCollection.FindOne(ctx, bson.M{"task_id":taskId}).Decode(&task)
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		workerTask.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		workerTask.ID = primitive.NewObjectID()
		workerTask.Worker_task_id = workerTask.ID.Hex()
		workerTask.Worker_id = uid
		workerTask.Task_id = taskId


		if *task.Approve_on_demand == true{
			workerTask.Last_status = "WAITING"
			taskStatus.Status = "WAITING"
		}else{
			workerTask.Last_status = "DOING"
			taskStatus.Status = "DOING"
		}
		taskStatus.Worker_task_id = workerTask.Worker_task_id
		taskStatus.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		taskStatus.ID = primitive.NewObjectID()
		taskStatus.Status_id = taskStatus.ID.Hex()

		//Insert taskStatus into DB---------------------------------
		_, insertErr := taskStatusCollection.InsertOne(ctx, taskStatus)
		if insertErr !=nil {
			msg := fmt.Sprintf("taskStatus item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
			return
		}
		defer cancel()

		//Insert workerTask into DB---------------------------------
		resultInsertionNumber, insertErr := workerTaskCollection.InsertOne(ctx, workerTask)
		if insertErr !=nil {
			msg := fmt.Sprintf("workerTask item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)


	}
}

//choose between doing and waiting ---> BUG : when there is nothing it panic
func GetWorkerTasks()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		last_status := c.Param("last_status")

		matchStage := bson.D{}
		matchStage2 := bson.D{}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage <1{
			recordPerPage = 10000
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 !=nil || page<1{
			page = 1
		}
		startIndex := (page - 1) * recordPerPage
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}}, 
			{"total_count", bson.D{{"$sum", 1}}}, 
			{"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}

		if UserRole == "ADMIN" {
			matchStage = bson.D{{"$match",bson.M{}}}
		}else if UserRole == "REQUESTER" {
			matchStage := bson.D{
				{"$lookup",
				bson.M{
				"from": "task",
				"localField": "task_id",
				"foreignField": "task_id",
				"as": "alltasks",
				}},

			}
			if last_status == "DOING" || last_status == "WAITING" {
				matchStage2 = bson.D{
					{"$match",bson.D{
						{"$and",
							bson.A{
								bson.D{{"alltasks.requester_id", uid}},
								bson.D{{"last_status", last_status}},
	
	
							}},
					},
					},
				}
			}else{
				matchStage2 = bson.D{{"$match",bson.M{"alltasks.requester_id": uid}}}
				// matchStage2 = bson.D{
				// 	{"$match",bson.D{
				// 		{"$or",
				// 			bson.A{
				// 				bson.D{{"alltasks.requester_id", uid}},	
				// 			}},
				// 	},
				// 	},
				// }
			}

			
			result,err := workerTaskCollection.Aggregate(ctx, mongo.Pipeline{
				matchStage,matchStage2, groupStage, projectStage})

			defer cancel()
			if err!=nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while listing user items"})
			}
			var allusers []bson.M
			if err = result.All(ctx, &allusers); err!=nil{
				log.Fatal(err)
			}
			c.JSON(http.StatusOK, allusers[0])
		}else if UserRole == "WORKER" {
			if last_status == "DOING" || last_status == "WAITING" {
				matchStage = bson.D{
					{"$match",bson.D{
						{"$and",
							bson.A{
								bson.D{{"worker_id", uid}},
								bson.D{{"last_status", last_status}},
	
							}},
					},
					},
				}
			}else{
				matchStage = bson.D{
					{"$match",bson.D{
						{"$or",
							bson.A{
								bson.D{{"worker_id", uid}},
							}},
					},
					},
				}
			}

			result,err := workerTaskCollection.Aggregate(ctx, mongo.Pipeline{
				matchStage, groupStage, projectStage})
			defer cancel()
			if err!=nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while listing user items"})
			}
			var allWorkerTasks []bson.M
			if err = result.All(ctx, &allWorkerTasks); err!=nil{
				log.Fatal(err)
			}
			c.JSON(http.StatusOK, allWorkerTasks[0])
		}

		
	}
}

//accept or reject application by requester
func EvaluateWorkerTask()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		workerTaskId := c.Param("worker_task_id")

		var task models.Task
		var workerTask models.WorkerTask
		var workerTaskEval models.WorkerTaskEval
		var taskStatus models.TaskStatus

		if UserRole == "WORKER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized to access this resource"})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		err := workerTaskCollection.FindOne(ctx, bson.M{"worker_task_id":workerTaskId}).Decode(&workerTask)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = taskCollection.FindOne(ctx, bson.M{"task_id":workerTask.Task_id}).Decode(&task)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if task.Requester_id != uid && UserRole == "REQUESTER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error":" no documents in result"})
			return
		}
		if err := c.BindJSON(&workerTaskEval); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		workerTaskEval.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		validationErr := validate1.Struct(workerTaskEval)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		}
		upsert := true
		filter := bson.M{"worker_task_id":workerTaskId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		_, err = workerTaskCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", workerTaskEval},
			},
			&opt,
		)
		defer cancel()
		if err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		taskStatus.Status = workerTaskEval.Last_status
		taskStatus.Worker_task_id = workerTask.Worker_task_id
		taskStatus.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		taskStatus.ID = primitive.NewObjectID()
		taskStatus.Status_id = taskStatus.ID.Hex()

		//Insert taskStatus into DB---------------------------------
		_, insertErr := taskStatusCollection.InsertOne(ctx, taskStatus)
		if insertErr !=nil {
			msg := fmt.Sprintf("taskStatus item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, workerTaskEval)

	}
}



