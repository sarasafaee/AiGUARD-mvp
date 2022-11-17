package models

import(
	"time"
"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct{
	ID					primitive.ObjectID		`bson:"_id"`
	Max_workers			*int					`json:"max_workers" validate:"required"`
	Title 				*string					`json:"title" validate:"required"`
	Description			*string					`json:"description`
	Approve_on_demand	*bool					`json:"approve_on_demand" validate:"required"`
	Deadline			time.Time				`json:"deadline" validate:"required"` //it seems it has a value even when its empty --> ??
	Wage				*string					`json:"wage`
	Stream_id			string					`json:"stream_id" validate:"required"`
	Requester_id		string					`json:"requester_id"`
	Created_at			time.Time				`json:"created_at"`
	Updated_at			time.Time				`json:"updated_at"`
	Task_id				string					`json:"task_id"`
}

type FilterTask struct{
	ID					primitive.ObjectID		`bson:"_id"`
	Tags				[]string				`json:"tags"`
	Created_at			time.Time				`json:"created_at"`
	Updated_at			time.Time				`json:"updated_at"`
	Task_id				string					`json:"task_id" validate:"required"`
	Filter_task_id		string 					`json:"filter_task_id" `
}

type WorkerTask struct{
	ID					primitive.ObjectID		`bson:"_id"`
	Created_at			time.Time				`json:"created_at"`
	Updated_at			time.Time				`json:"updated_at"`
	Task_id				string					`json:"task_id" validate:"required"`
	Worker_id			string 					`json:"filter_task_id"`
	Worker_task_id		string					`json:"worker_task_id"`
	Last_status			string					`json:"last_status" validate:"required,eq=PAID|eq=DONE|eq=DOING|eq=REJECTED|eq=WAITING|eq=SUSPENDED"`
}

//to use for evluation of workertask
type WorkerTaskEval struct{
	Updated_at			time.Time				`json:"updated_at"`
	Last_status			string					`json:"last_status" validate:"required,eq=PAID|eq=DONE|eq=DOING|eq=REJECTED|eq=WAITING|eq=SUSPENDED"`
}

type TaskStatus struct{
	ID					primitive.ObjectID		`bson:"_id"`
	Created_at			time.Time				`json:"created_at"`
	Worker_task_id		string					`json:"worker_task_id"`
	Status_id			string					`json:"status_id"`
	Status				string					`json:"last_status" validate:"required,eq=PAID|eq=DONE|eq=DOING|eq=REJECTED|eq=WAITING|eq=SUSPENDED"`				
}
