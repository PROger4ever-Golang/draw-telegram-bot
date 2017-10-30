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
	FirstName        string
	LastName         string
	Status           string
}

func (m *User) Init(collection *UserCollection) *User {
	m.BaseModel = mongo.NewModel(collection.BaseCollection, m)
	return m
}

func (m *User) GetBaseModel() *mongo.BaseModel {
	return m.BaseModel
}

func (m *User) SetBaseModel(bm *mongo.BaseModel) {
	m.BaseModel = bm
}

func (m *User) CleanModel() {
	m.TelegramID = 0
	m.Username = ""
	m.FirstName = ""
	m.LastName = ""
	m.Status = ""
}

func (m *User) GetContentMap() bson.M {
	return bson.M{
		"telegram_id": m.TelegramID,
		"username":    m.Username,
		"firstname":   m.FirstName,
		"lastname":    m.LastName,
		"status":      m.Status,
	}
}

func (m *User) SetContentFromMap(theMap bson.M) {
	if telegramIDI, okM := theMap["telegram_id"]; okM {
		if telegramID, okC := telegramIDI.(int); okC {
			m.TelegramID = telegramID
		}
	}

	if UsernameI, okM := theMap["username"]; okM {
		if Username, okC := UsernameI.(string); okC {
			m.Username = Username
		}
	}

	if FirstNameI, okM := theMap["firstname"]; okM {
		if FirstName, okC := FirstNameI.(string); okC {
			m.FirstName = FirstName
		}
	}

	if LastNameI, okM := theMap["lastname"]; okM {
		if LastName, okC := LastNameI.(string); okC {
			m.LastName = LastName
		}
	}

	if StatusI, okM := theMap["status"]; okM {
		if Status, okC := StatusI.(string); okC {
			m.Status = Status
		}
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
	obj.Init(c)
	err = c.BaseCollection.FindOneModel(query, obj)
	return
}

func (c *UserCollection) Insert(values ...*User) error {
	models := c.toModels(values)
	return c.BaseCollection.InsertModel(models...)
}

func (c *UserCollection) InsertOneOrUpdateModel(value *User) (isUpdated bool, err error) {
	return c.BaseCollection.InsertOneOrUpdateModel(bson.M{
		"telegram_id": value.TelegramID,
	}, value)
}

func (c *UserCollection) PipeOne(pipeline interface{}) (obj *User, err error) {
	obj = &User{}
	obj.Init(c)
	err = c.BaseCollection.PipeOneModel(pipeline, obj)
	return
}

func (c *UserCollection) toModels(values []*User) (models []mongo.Model) {
	models = make([]mongo.Model, 0, len(values))
	for _, v := range values {
		models = append(models, v)
	}
	return
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
