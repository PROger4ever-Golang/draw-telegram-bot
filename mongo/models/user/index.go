package user

import (
	"reflect"

	"bitbucket.org/proger4ever/draw-telegram-bot/mongo"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	*mongo.BaseModel `bson:"-"`
	TelegramID       int `bson:"telegram_id"`
	Username         string
	Roles            []string
	FirstName        string
	LastName         string
	Status           string
}

func (m *User) Init(collection *UserCollection) *User {
	m.BaseModel = mongo.NewModel(collection.Collection, m)
	return m
}

func (m *User) GetContentMap() bson.M {
	return bson.M{
		"telegram_id": m.TelegramID,
		"username":    m.Username,
		"roles":       m.Roles,
		"firstname":   m.FirstName,
		"lastname":    m.LastName,
		"status":      m.Status,
	}
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

func (c *UserCollection) FindOne(query bson.M) (obj *User, err error) {
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

func NewCollectionDefault() *UserCollection {
	c := UserCollection{}
	return c.Init(mongo.DefaultConnection)
}
