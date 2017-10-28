package user

import (
	"reflect"

	"gopkg.in/mgo.v2"

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
	m.BaseModel = mongo.NewModel(collection.BaseCollection, m)
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
	*mongo.BaseCollection
}

func (c *UserCollection) Init(connection *mongo.Connection) {
	c.BaseCollection = mongo.NewCollection(connection, c, "mazimotaBot", "users", reflect.TypeOf(&User{}))
}

func (c *UserCollection) GetIndexes() []mgo.Index {
	return []mgo.Index{{
		Key:        []string{"telegram_id"},
		Background: false,
		DropDups:   true,
		Unique:     true,
	}}
}

func (c *UserCollection) EnsureIndexes() error {
	return c.BaseCollection.EnsureIndexes(c.GetIndexes())
}

func (c *UserCollection) FindOne(query bson.M) (obj *User, err error) {
	obj = &User{}
	err = c.BaseCollection.FindOneUnsafe(query, obj)
	return
}

func (c *UserCollection) Insert(values ...*User) error {
	return c.BaseCollection.InsertUnsafe(values)
}

func New(collection *UserCollection) *User {
	m := User{}
	return m.Init(collection)
}

func NewCollection(connection *mongo.Connection) (c *UserCollection) {
	c = &UserCollection{}
	c.Init(connection)
	return
}

func NewCollectionDefault() (c *UserCollection) {
	c = &UserCollection{}
	c.Init(mongo.DefaultConnection)
	return
}
