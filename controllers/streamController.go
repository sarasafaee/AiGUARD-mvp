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
helper "github.com/sarasafaee/AiGUARD-mvp/helpers"
"github.com/sarasafaee/AiGUARD-mvp/models"
"github.com/sarasafaee/AiGUARD-mvp/database"

"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/bson/primitive"
"go.mongodb.org/mongo-driver/mongo"
"go.mongodb.org/mongo-driver/mongo/options"
)

var streamCollection *mongo.Collection = database.OpenCollection(database.Client, "stream")
var activityStreamCollection *mongo.Collection = database.OpenCollection(database.Client, "activityStream")



var validate2 = validator.New()

func CtreateStream()gin.HandlerFunc{

	return func(c *gin.Context){
		uid := c.GetString("uid")
		//check access---------------
		if err := helper.CheckUserRole(c, "REQUESTER"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		//----------------------------
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var stream models.Stream

		if err := c.BindJSON(&stream); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		stream.State = "ALIVE"
		validationErr := validate2.Struct(stream)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
		} else {
			//stream Existence in DB -----------------------
			streamCount, err := streamCollection.CountDocuments(ctx, bson.M{"stream_id":stream.Stream_id,"state":"ALIVE"})
			defer cancel()
			if err != nil {
				log.Panic(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while checking for the stream"})
				return
			}

			if streamCount >0{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"stream already exists"})
			}else{
				stream.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				stream.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				stream.ID = primitive.NewObjectID()
				stream.Stream_id = stream.ID.Hex()
				stream.Requester_id = uid
				//Insert stream into DB---------------------------------
				resultInsertionNumber, insertErr := streamCollection.InsertOne(ctx, stream)
				if insertErr !=nil {
					msg := fmt.Sprintf("Stream item was not created")
					c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
					return
				}
				defer cancel()
				c.JSON(http.StatusOK, resultInsertionNumber)
			}
	
		}

	}

}

//get a stream just for requesterowner and admin
func GetStream()gin.HandlerFunc{

	return func(c *gin.Context){
		
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		streamId := c.Param("stream_id")
		var stream models.Stream
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		if UserRole == "WORKER" {
			c.JSON(http.StatusBadRequest, gin.H{"error" : "Unauthorized to access this resource"})
			return
		}else if UserRole == "REQUESTER" {
			//check stream existence
			err := streamCollection.FindOne(ctx, bson.M{"stream_id":streamId,"state":"ALIVE"}).Decode(&stream)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}else if UserRole == "ADMIN"{
			//check stream existence
			err := streamCollection.FindOne(ctx, bson.M{"stream_id":streamId}).Decode(&stream)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		//check stream belonging to owner or is an admin
		if stream.Requester_id == uid || UserRole == "ADMIN"{
			c.JSON(http.StatusOK, stream)
			return
		}else{
			c.JSON(http.StatusInternalServerError, gin.H{"error": " no documents in result"})
			return
		}
	
		
	}
}

//get all streams just for requesterowner and admin
func GetStreams()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		matchStage := bson.D{}
		if UserRole == "ADMIN" {
			matchStage = bson.D{{"$match",bson.M{}}}

		}else if UserRole == "REQUESTER" {
			matchStage = bson.D{
				{"$match",bson.D{
					{"$and",
						bson.A{
							bson.D{{"state", "ALIVE"}},
							bson.D{{"requester_id", uid}},
						}},
				},
				},
			}
			// matchStage = bson.D{{"$match",bson.M{"requester_id":uid}}}
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
				result,err := streamCollection.Aggregate(ctx, mongo.Pipeline{
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

//edit a stream by requesterowner and admin
func EditStream()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")

		var stream models.Stream
		if UserRole == "WORKER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized to access this resource"})
			return
		}
		streamId := c.Param("stream_id")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		err := streamCollection.FindOne(ctx, bson.M{"stream_id":streamId}).Decode(&stream)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if stream.Requester_id != uid && UserRole == "REQUESTER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error":" no documents in result"})
			return
		}
		if err := c.BindJSON(&stream); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		stream.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		validationErr := validate1.Struct(stream)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		}
		upsert := true
		filter := bson.M{"stream_id":streamId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
	
		_, err = streamCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", stream},
			},
			&opt,
		)
	
		defer cancel()
	
		if err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, stream)


	}
}

//delete a stream and all its dependencies(task,stream_activity) by requesterowner and admin 
func DeleteStream()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")

		var stream models.Stream
		if UserRole == "WORKER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized to access this resource"})
			return
		}
		streamId := c.Param("stream_id")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		err := streamCollection.FindOne(ctx, bson.M{"stream_id":streamId}).Decode(&stream)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if UserRole == "REQUESTER"{
			if stream.Requester_id != uid {
				c.JSON(http.StatusInternalServerError, gin.H{"error":" no documents in result"})
				return
			}	
		}

		_, err = streamCollection.DeleteOne(ctx, bson.M{"stream_id":streamId})
		var task models.Task
		_ = taskCollection.FindOne(ctx, bson.M{"stream_id":streamId}).Decode(&task)//might cause BUG
		_, err = taskCollection.DeleteOne(ctx, bson.M{"stream_id":streamId})//might cause BUG
		_, err = filterTaskCollection.DeleteOne(ctx, bson.M{"task_id":task.Task_id})//might cause BUG

		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while deleting task"})
		}
		c.JSON(http.StatusOK,gin.H{"message":"stream deleted successfully"} )
	}
}

//create a activityStream by requester
func CreateActivityStream()gin.HandlerFunc{
	return func(c *gin.Context){
		uid := c.GetString("uid")
		//check access---------------
		if err := helper.CheckUserRole(c, "REQUESTER"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		//----------------------------
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var activityStream models.ActivityStream

		if err := c.BindJSON(&activityStream); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate2.Struct(activityStream)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		} else {
			//stream Existence in DB -----------------------

			var stream models.Stream
			err := streamCollection.FindOne(ctx, bson.M{"stream_id":activityStream.Stream_id}).Decode(&stream)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": " stream not found"})
				return
			}
			//check stream belonging to owner
			if stream.Requester_id != uid {
				c.JSON(http.StatusInternalServerError, gin.H{"error": " stream not found"})
				return
			}
			//action Existence in DB -----------------------
			var action models.Action
			err = actionCollection.FindOne(ctx, bson.M{"action_id":activityStream.Action_id}).Decode(&action)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": " action not found"})
				return
			}
			activityStream.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			activityStream.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			activityStream.ID = primitive.NewObjectID()
			activityStream.Activity_stream_id = activityStream.ID.Hex()
			//Insert activityStream into DB---------------------------------
			resultInsertionNumber, insertErr := activityStreamCollection.InsertOne(ctx, activityStream)
			if insertErr !=nil {
				msg := fmt.Sprintf("activityStream item was not created")
				c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
				return
			}
			defer cancel()
			c.JSON(http.StatusOK, resultInsertionNumber)
	
		}

	}
}

//get activityStream By activityStreamId
func GetActivityStreamByID()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		activityStreamId := c.Param("activity_stream_id")
		var activityStream models.ActivityStream
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		//check activityStream existence
		err := activityStreamCollection.FindOne(ctx, bson.M{"activity_stream_id":activityStreamId}).Decode(&activityStream)
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if UserRole == "WORKER" {
			c.JSON(http.StatusBadRequest, gin.H{"error" : "Unauthorized to access this resource"})
			return
		}else if UserRole == "REQUESTER" {
			//check activityStream and task belonging to owner
			var stream models.Stream
			err := streamCollection.FindOne(ctx, bson.M{"stream_id":activityStream.Stream_id}).Decode(&stream)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if stream.Requester_id == uid {
				c.JSON(http.StatusOK, activityStream)
				return
			}else{
				c.JSON(http.StatusInternalServerError, gin.H{"error": " no documents in result"})
				return
			}
		}else if UserRole == "ADMIN"{
			c.JSON(http.StatusOK, activityStream)

		}
		
	}
}

//get activityStreams By streamId
func GetActivityStreamsByStreamID()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		streamId := c.Param("stream_id")

		matchStage := bson.D{}
		if UserRole == "ADMIN" {
			matchStage = bson.D{{"$match",bson.M{"stream_id":streamId}}}

		}else if UserRole == "REQUESTER" {
			var stream models.Stream
			var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
			//check stream existence
			err := streamCollection.FindOne(ctx, bson.M{"stream_id":streamId}).Decode(&stream)
			defer cancel()
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if stream.Requester_id != uid {
				c.JSON(http.StatusInternalServerError,  gin.H{"error": " no documents in result"})
				return
			}
			matchStage = bson.D{{"$match",bson.M{"stream_id":streamId}}}
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
				result,err := activityStreamCollection.Aggregate(ctx, mongo.Pipeline{
					matchStage, groupStage, projectStage})
				defer cancel()
				if err!=nil{
					c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while listing user items"})
				}
				var allactivities []bson.M
				if err = result.All(ctx, &allactivities); err!=nil{
					log.Fatal(err)
				}
				c.JSON(http.StatusOK, allactivities[0])
	}
}

//edit a activityStream by requesterowner and admin
func EditActivityStream()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")

		var activityStream models.ActivityStream
		if UserRole == "WORKER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized to access this resource"})
			return
		}
		activityStreamId := c.Param("activity_stream_id")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		err := activityStreamCollection.FindOne(ctx, bson.M{"activity_stream_id":activityStreamId}).Decode(&activityStream)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var stream models.Stream

		if UserRole == "REQUESTER"{
			err := streamCollection.FindOne(ctx, bson.M{"stream_id":activityStream.Stream_id}).Decode(&stream)
			defer cancel()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if stream.Requester_id != uid {
				c.JSON(http.StatusInternalServerError, gin.H{"error":" no documents in result"})
				return
			}

			
		}
		if err := c.BindJSON(&activityStream); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//update updated_at --------------------------
		activityStream.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		validationErr := validate1.Struct(activityStream)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		}
		//costum validation 
		if activityStream.Stream_id != stream.Stream_id || activityStreamId != activityStream.Activity_stream_id{
			c.JSON(http.StatusBadRequest, gin.H{"error":"editing StreamId and activityStreamId is not possible"})
			return
		} 
		var action models.Action
		err = actionCollection.FindOne(ctx, bson.M{"action_id":activityStream.Action_id}).Decode(&action)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "action not found"})
			return
		}
		// ---------------------------------------

		upsert := true
		filter := bson.M{"activity_stream_id":activityStreamId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		_, err = activityStreamCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", activityStream},
			},
			&opt,
		)
	
		defer cancel()
		if err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, activityStream)


	}
}

//delete a activityStream by requesterowner and admin
func DeleteActivityStream()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")

		var activityStream models.ActivityStream
		if UserRole == "WORKER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized to access this resource"})
			return
		}
		activityStreamId := c.Param("activity_stream_id")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		err := activityStreamCollection.FindOne(ctx, bson.M{"activity_stream_id":activityStreamId}).Decode(&activityStream)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var stream models.Stream

		if UserRole == "REQUESTER"{
			err := streamCollection.FindOne(ctx, bson.M{"stream_id":activityStream.Stream_id}).Decode(&stream)
			defer cancel()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if stream.Requester_id != uid {
				c.JSON(http.StatusInternalServerError, gin.H{"error":" no documents in result"})
				return
			}	
		}
		_, err = activityStreamCollection.DeleteOne(ctx, bson.M{"activity_stream_id":activityStreamId})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while deleting activity"})
		}
		c.JSON(http.StatusOK,gin.H{"message":"activity deleted successfully"} )


	}
}
