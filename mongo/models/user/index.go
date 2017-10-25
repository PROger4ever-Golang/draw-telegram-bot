package user

import (
	"reflect"
	"time"

	"bitbucket.org/proger4ever/draw-telegram-bot/mongo"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	*mongo.Model `bson:"-"`
	TelegramID   int `bson:"telegram_id"`
	Username     string
	Roles        []string
	FirstName    string
	LastName     string
	Status       string
	CreatedAt    time.Time `bson:"created_at,omitempty"`
	UpdatedAt    time.Time `bson:"updated_at,omitempty"`
	DeletedAt    time.Time `bson:"deleted_at,omitempty"`
}

func (c *User) Init(collection *UserCollection) *User {
	c.Model = mongo.NewModel(collection.Collection, c)
	return c
}

type UserCollection struct {
	*mongo.Collection
	MongoSession   *mongo.Connection
	DbName         string
	CollectionName string
}

func (c *UserCollection) Init(connection *mongo.Connection) *UserCollection {
	c.Collection = mongo.NewCollection(connection, "mazimotaBot", "users", reflect.TypeOf(&User{}))
	return c
}

func (c *UserCollection) FindOne(query *bson.M) (obj *User, err error) {
	obj = &User{}
	err = c.Collection.FindOneUnsafe(query, obj)
	return
}

func (c *UserCollection) Insert(values ...*User) error {
	return c.Collection.InsertUnsafe(values)
}

func New(collection *UserCollection) *User {
	m := User{}
	return m.Init(collection)
}

func NewCollection(connection *mongo.Connection) *UserCollection {
	c := UserCollection{}
	return c.Init(connection)
}
