package helper

import(
	"errors"
	"github.com/gin-gonic/gin"
)


func CheckUserRole(c *gin.Context, role string) (err error){
	UserRole := c.GetString("User_role")
	err = nil
	if UserRole != role {
		err = errors.New("Unauthorized to access this resource")
		return err
	}
	return err
}

func MatchUserRoleToUid(c *gin.Context, userId string) (err error){
	UserRole := c.GetString("User_role")
	uid := c.GetString("uid")
	err= nil

	if UserRole == "REQUESTER"  && uid != userId {
		err = errors.New("Unauthorized to access this resource")
		return err
	}else if UserRole == "WORKER"  && uid != userId {
		err = errors.New("Unauthorized to access this resource")
		return err
	}
	err = CheckUserRole(c, UserRole)
	return err
}