package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Email        string             `json:"email" bson:"email"`
	Password     string             `json:"password" bson:"password"`
	CreationTime time.Time          `json:"creationDate" bson:"creationDate"`
	Student      bool               `json:"student" bson:"student"`
	Ayudante     bool               `json:"ayudante" bson:"ayudante"`
}

type BlogTheme struct {
	IdTheme      primitive.ObjectID `json:"idBlog" bson:"idBlog"`
	Theme        string             `json:"theme" bson:"theme"`
	Description  string             `json:"description" bson:"description"`
	NameCreator  string             `json:"nameCreator" bson:"nameCreator"`
	IdCreator    primitive.ObjectID `json:"idCreator" bson:"idCreator"`
	CreationTime time.Time          `json:"creationDate" bson:"creationDate"`
}

type BlogComment struct {
	IdComment    primitive.ObjectID `json:"idComment" bson:"idComment"`
	IdTheme      primitive.ObjectID `json:"idTheme" bson:"idTheme"`
	Description  string             `json:"description" bson:"description"`
	Latex        string             `json:"latex" bson:"latex"`
	NameCreator  string             `json:"nameCreator" bson:"nameCreator"`
	IdCreator    primitive.ObjectID `json:"idCreator" bson:"idCreator"`
	CreationTime time.Time          `json:"creationDate" bson:"creationDate"`
}
