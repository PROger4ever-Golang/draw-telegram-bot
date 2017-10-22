package user

import (
	"time"

	"bitbucket.org/proger4ever/draw-telegram-bot/mongo"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/base-model"
)

type User struct {
	ID        int `bson:"_id,omitempty"`
	Username  string
	Type      string
	FirstName string
	LastLame  string
	Status    string
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	DeletedAt time.Time `bson:"deleted_at"`
}

type UserModel struct {
	*baseModel.BaseModel
	MongoSession   *mongo.Connection
	DbName         string
	CollectionName string
}

func (m *UserModel) Init(connection *mongo.Connection, collectionName string) {
	m.BaseModel = baseModel.NewBaseModel(connection, "mazimotaBot", "users")
}

func NewUserModel(connection *mongo.Connection) *UserModel {
	m := UserModel{}
	m.Init(connection)
	return &m
}
