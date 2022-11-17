package models

import(
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Action struct{
	ID					primitive.ObjectID		`bson:"_id"`
	Name 				*string					`json:"name" validate:"required"`
	Requester_id		string					`json:"requester_id"`
	Created_at			time.Time				`json:"created_at"`
	Updated_at			time.Time				`json:"updated_at"`
	Action_id			string					`json:"action_id"`
}