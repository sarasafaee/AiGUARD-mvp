package middleware

import(
	"fmt"
	"net/http"
	helper "github.com/sarasafaee/AiGUARD-mvp/helpers"
	"github.com/gin-gonic/gin"
)

func NotificationSendingAuthenticate() gin.HandlerFunc{
	return func(c *gin.Context){
		clientToken := c.Request.Header.Get("token")
		if clientToken == ""{
			c.JSON(http.StatusInternalServerError, gin.H{"error":fmt.Sprintf("No Authorization header provided")})
			c.Abort()
			return
		}

		claims, err := helper.ValidateNotificationToken(clientToken)
		if err !="" {
			c.JSON(http.StatusInternalServerError, gin.H{"error":err})
			c.Abort()
			return
		}
		c.Set("stream_id", claims.StreamID)
		c.Set("task_id", claims.TaskID)
		c.Set("notification_token_id", claims.NotificationTokenID)
		c.Set("uid",claims.Uid)
		c.Next()
	}
}