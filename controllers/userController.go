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
// "golang.org/x/crypto/bcrypt"

"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/bson/primitive"
"go.mongodb.org/mongo-driver/mongo"
"go.mongodb.org/mongo-driver/mongo/options"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var blacklistTokenCollection *mongo.Collection = database.OpenCollection(database.Client, "blacklistToken")

var validate = validator.New()

// func HashPassword(password string) string{
// 	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
// 	if err!=nil{
// 		log.Panic(err)
// 	}
// 	return string(bytes)
// }

// func VerifyPassword(userPassword string, providedPassword string)(bool, string){
// 	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
// 	check := true
// 	msg := ""

// 	if err!= nil {
// 		msg = fmt.Sprintf("email or password is incorrect")
// 		check=false
// 	}
// 	return check, msg
// }

func Signup()gin.HandlerFunc{

	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return
		} else {
			emailCount, err := userCollection.CountDocuments(ctx, bson.M{"email":user.Email,"state":"ALIVE"})
			defer cancel()
			if err != nil {
				log.Panic(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while checking for the email"})
			}
	
			// password := HashPassword(*user.Password)
			// user.Password = &password
	
			phoneCount, err := userCollection.CountDocuments(ctx, bson.M{"phone":user.Phone,"state":"ALIVE"})
			defer cancel()
			if err!= nil {
				log.Panic(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while checking for the phone number"})
			}
	
			if phoneCount >0{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"this phone number already exists"})
			}else if emailCount >0{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"this email already exists"})
			}else{
				user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				user.ID = primitive.NewObjectID()
				user.User_id = user.ID.Hex()
				token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_role, *&user.User_id)
				user.Token = &token
				user.Refresh_token = &refreshToken
				user.State = "ALIVE"
		
				resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
				if insertErr !=nil {
					msg := fmt.Sprintf("User item was not created")
					c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
					return
				}
				defer cancel()
				c.JSON(http.StatusOK, resultInsertionNumber)
			}
	
		}



	}

}

func Login() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return 
		}
		if user.Password == nil || user.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"email and password are required"})
		}else{
			err := userCollection.FindOne(ctx, bson.M{"email":user.Email,"state":"ALIVE"}).Decode(&foundUser)
			defer cancel()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error":"email or password is incorrect"})
				return
			}
	
			// passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
			if *user.Password != *foundUser.Password {
				c.JSON(http.StatusInternalServerError, gin.H{"error":"email or password is incorrect"})
				return
			}
			defer cancel()
			// if passwordIsValid != true{
			// 	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			// 	return
			// }
	
			if foundUser.Email == nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"user not found"})
			}
			token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_role, foundUser.User_id)
			helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
			err = userCollection.FindOne(ctx, bson.M{"user_id":foundUser.User_id}).Decode(&foundUser)
			// c.Header("token", token)
			// c.SetCookie(
			// 	"token",
			// token,
			// 60*60*24,
			// "",
			// "",
			// false,
			// false,
			// )
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, foundUser)
		}

	}
}

func Logout() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var blacklistToken models.BlacklistToken
		token := c.GetHeader("token")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "Authorization Token is required"})
			c.Abort()
			return
		}
		blacklistToken.Token = &token
		tokenCount, err := blacklistTokenCollection.CountDocuments(ctx, bson.M{"token":blacklistToken.Token})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while checking for the token"})
		}
		if tokenCount >0{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"you already logedOut"})
		}else {
			resultInsertionNumber, insertErr := blacklistTokenCollection.InsertOne(ctx, blacklistToken)
			if insertErr !=nil {
				msg := fmt.Sprintf("blacklistuser item was not created")
				c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
				return
			}
			defer cancel()
			c.JSON(http.StatusOK, resultInsertionNumber)
		}

	}
}

func GetUsers() gin.HandlerFunc{
	return func(c *gin.Context){
		if err := helper.CheckUserRole(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
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

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}}, 
			{"total_count", bson.D{{"$sum", 1}}}, 
			{"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},}}}
				result,err := userCollection.Aggregate(ctx, mongo.Pipeline{
					matchStage, groupStage, projectStage})
				defer cancel()
				if err!=nil{
					c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while listing user items"})
				}
				var allusers []bson.M
				if err = result.All(ctx, &allusers); err!=nil{
					log.Fatal(err)
				}
				c.JSON(http.StatusOK, allusers[0])}
}

func GetUser() gin.HandlerFunc{
	return func(c *gin.Context){
		userId := c.Param("user_id")

		if err := helper.MatchUserRoleToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id":userId}).Decode(&user)
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func EditUser() gin.HandlerFunc{
	return func(c *gin.Context){
		var changedUser models.User
		var originalUser models.User

		userId := c.Param("user_id")

		if err := helper.MatchUserRoleToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		err := userCollection.FindOne(ctx, bson.M{"user_id":userId}).Decode(&changedUser)
		err = userCollection.FindOne(ctx, bson.M{"user_id":userId}).Decode(&originalUser)

		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := c.BindJSON(&changedUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if *changedUser.Password != *originalUser.Password {
			// password := HashPassword(*changedUser.Password)
			password := *changedUser.Password
			changedUser.Password = &password
		} 

		changedUser.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		validationErr := validate.Struct(changedUser)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			
		}else {
			upsert := true
			filter := bson.M{"user_id":userId}
			opt := options.UpdateOptions{
				Upsert: &upsert,
			}
		
			_, err = userCollection.UpdateOne(
				ctx,
				filter,
				bson.D{
					{"$set", changedUser},
				},
				&opt,
			)
		
			defer cancel()
		
			if err!=nil{
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}
			c.JSON(http.StatusOK, changedUser)
		}


	}
}

//deactivateMyself
func DeactivateUser() gin.HandlerFunc{
	return func(c *gin.Context){
		token := c.GetHeader("token")
		uid := c.GetString("uid")

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		// deleteResult, err := userCollection.DeleteOne(ctx, bson.M{"user_id":uid})
		// defer cancel()
		// if err != nil {
		// 	log.Panic(err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while deleting user"})
		// }
		var changedUser models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id":uid}).Decode(&changedUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		changedUser.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		changedUser.State = "DELETED"
		upsert := true
		filter := bson.M{"user_id": uid}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
	
		_, err = userCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", changedUser},
			},
			&opt,
		)
		defer cancel()
		
		if err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var blacklistToken models.BlacklistToken
		blacklistToken.Token = &token

		_, insertErr := blacklistTokenCollection.InsertOne(ctx, blacklistToken)
		if insertErr !=nil {
			msg := fmt.Sprintf("blacklistuser item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, gin.H{"message":uid})
	}
}

