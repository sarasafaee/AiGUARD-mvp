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

var actionCollection *mongo.Collection = database.OpenCollection(database.Client, "action")



var validate3 = validator.New()

func CtreateAction()gin.HandlerFunc{

	return func(c *gin.Context){
		uid := c.GetString("uid")
		//check access---------------
		if err := helper.CheckUserRole(c, "REQUESTER"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		//----------------------------
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var action models.Action

		if err := c.BindJSON(&action); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate3.Struct(action)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
		} else {
			//action Existence in DB -----------------------
			actionCount, err := actionCollection.CountDocuments(ctx, bson.M{"action_id":action.Action_id})
			defer cancel()
			if err != nil {
				log.Panic(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while checking for the action"})
				return
			}

			if actionCount >0{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"action already exists"})
			}else{
				action.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				action.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				action.ID = primitive.NewObjectID()
				action.Action_id = action.ID.Hex()
				action.Requester_id = uid
				//Insert action into DB---------------------------------
				resultInsertionNumber, insertErr := actionCollection.InsertOne(ctx, action)
				if insertErr !=nil {
					msg := fmt.Sprintf("Action item was not created")
					c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
					return
				}
				defer cancel()
				c.JSON(http.StatusOK, resultInsertionNumber)
			}
	
		}

	}

}

//get a action just for requesterowner and admin
func GetAction()gin.HandlerFunc{

	return func(c *gin.Context){
		
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		actionId := c.Param("action_id")
		var action models.Action
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		if UserRole == "WORKER" {
			c.JSON(http.StatusBadRequest, gin.H{"error" : "Unauthorized to access this resource"})
			return
		}
		//check action existence
		err := actionCollection.FindOne(ctx, bson.M{"action_id":actionId}).Decode(&action)
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		//check action belonging to owner or is an admin
		if action.Requester_id == uid || UserRole == "ADMIN"{
			c.JSON(http.StatusOK, action)
			return
		}else{
			c.JSON(http.StatusInternalServerError, gin.H{"error": " no documents in result"})
			return
		}
	
		
	}
}

//get all actions just for requesterowner and admin
func GetActions()gin.HandlerFunc{
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
				result,err := actionCollection.Aggregate(ctx, mongo.Pipeline{
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

//edit a action by requesterowner and admin
func EditAction()gin.HandlerFunc{
	return func(c *gin.Context){
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")

		var action models.Action
		if UserRole == "WORKER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unauthorized to access this resource"})
			return
		}
		actionId := c.Param("action_id")
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		err := actionCollection.FindOne(ctx, bson.M{"action_id":actionId}).Decode(&action)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if action.Requester_id != uid && UserRole == "REQUESTER"{
			c.JSON(http.StatusInternalServerError, gin.H{"error":" no documents in result"})
			return
		}
		if err := c.BindJSON(&action); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		action.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		validationErr := validate1.Struct(action)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		}
		upsert := true
		filter := bson.M{"action_id":actionId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
	
		_, err = actionCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", action},
			},
			&opt,
		)
	
		defer cancel()
	
		if err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, action)


	}
}

//delete a action by requesterowner or admin 
func DeleteAction()gin.HandlerFunc{
	return func(c *gin.Context){
				
		UserRole := c.GetString("User_role")
		uid := c.GetString("uid")
		actionId := c.Param("action_id")
		var action models.Action
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		if UserRole == "WORKER" {
			c.JSON(http.StatusBadRequest, gin.H{"error" : "Unauthorized to access this resource"})
			return
		}
		//check action existence
		err := actionCollection.FindOne(ctx, bson.M{"action_id":actionId}).Decode(&action)
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		//check action belonging to owner or is an admin
		if action.Requester_id == uid || UserRole == "ADMIN"{
			deleteResult, err := actionCollection.DeleteOne(ctx, bson.M{"action_id":actionId})
			defer cancel()
			if err != nil {
				log.Panic(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while deleting user"})
			}
			c.JSON(http.StatusOK, deleteResult)	
			return
		}else{
			c.JSON(http.StatusInternalServerError, gin.H{"error": " no documents in result"})
			return
		}
	}
}