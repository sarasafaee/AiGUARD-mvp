package models

import(
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct{
	ID						primitive.ObjectID		`bson:"_id"`
	Message 				 string					`json:"message" validate:"required"`
	Worker_task_id			 string					`json:"worker_task_id" validate:"required"`
	Created_at				 time.Time				`json:"created_at"`
	Updated_at				 time.Time				`json:"updated_at"`
	Notification_id			 string					`json:"Notification_id"`
	Last_status				 string					`json:"last_status" validate:"required,eq=DELIVERD|eq=SEEN|eq=SENT"`
	Notification_type		 string					`json:"notification_type" validate:"required,eq=ALARM|eq=REPORT"`
}

type NotificationToken struct{
	ID						primitive.ObjectID		`bson:"_id"`
	NotificationToken_id	string					`json:"notification_token_id"`
	Stream_id				string					`json:"stream_id" validate:"required"`
	Task_id					string					`json:"task_id"`
	Token					*string					`json:"token"`
}