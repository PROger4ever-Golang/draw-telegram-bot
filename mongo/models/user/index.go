package user

import (
	"reflect"
	"time"

	"bitbucket.org/proger4ever/draw-telegram-bot/mongo"
	"bitbucket.org/proger4ever/draw-telegram-bot/mongo/models/base-model"
	"gopkg.in/mgo.v2/bson"
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

func (m *UserModel) Init(connection *mongo.Connection) {
	m.BaseModel = baseModel.New(connection, "mazimotaBot", "users", reflect.TypeOf(&User{}))
}

func (m *UserModel) FindOne(query *bson.M) (obj *User, err error) {
	obj = &User{}
	err = m.BaseModel.FindOneUnsafe(query, obj)
	return
}

func (m *UserModel) Insert(values ...*User) error {
	return m.BaseModel.InsertUnsafe(values)
}

func New(connection *mongo.Connection) *UserModel {
	m := UserModel{}
	m.Init(connection)
	return &m
}
