package models

import(
	"time"
"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct{
	ID				primitive.ObjectID		`bson:"_id"`
	First_name		*string					`json:"first_name" validate:"required,min=2,max=100"`
	Last_name		*string					`json:"last_name" validate:"required,min=2,max=100"`
	Password		*string					`json:"Password" validate:"required,min=6"`
	Email			*string					`json:"email" validate:"email,required"`
	Phone			*string					`json:"phone" validate:"required"`
	Token			*string					`json:"token"`
	User_role		*string					`json:"User_role" validate:"required,eq=WORKER|eq=REQUESTER"`
	Refresh_token	*string					`json:"refresh_token"`
	Created_at		time.Time				`json:"created_at"`
	Updated_at		time.Time				`json:"updated_at"`
	User_id			string					`json:"user_id"`
	Location		*string					`json:"location" validate:"required,eq=IRAN|eq=IRAQ|eq=ENGLAND|eq=ITALY|eq=GERMANY|eq=NORWAY"`

}