package models

import(
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Stream struct{
	ID					primitive.ObjectID		`bson:"_id"`
	Title 				*string					`json:"title" validate:"required"`
	Requester_id		string					`json:"requester_id"`
	Created_at			time.Time				`json:"created_at"`
	Updated_at			time.Time				`json:"updated_at"`
	Token				*string					`json:"token  validate:"required"`
	Available_at		time.Time				`json:"available_at"` //search about time period type and ...
	Stream_id			string					`json:"stream_id"`
}

type FilterStream struct{
	ID					primitive.ObjectID		`bson:"_id"`
	Severity 			*int					`json:"severity" validate:"required"`
	Created_at			time.Time				`json:"created_at"`
	Updated_at			time.Time				`json:"updated_at"`
	Stream_id			string					`json:"stream_id" validate:"required"`
	Filter_stream_id 	string					`json:"filter_stream_id"`
	Action_id 			string 					`json:"action_id" validate:"required"`
}
