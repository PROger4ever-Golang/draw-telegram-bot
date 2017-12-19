package user

import (
	"reflect"

	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"gopkg.in/mgo.v2"

	"bitbucket.org/proger4ever/draw-telegram-bot/mongo"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type User struct {
	*mongo.BaseModel `bson:"-"`
	UserCollection   *UserCollection

	TelegramID int `bson:"telegram_id"`
	Username   string
	FirstName  string
	LastName   string
	Status     string

	LastAdditionAt time.Time `bson:"last_addition_at,omitempty"`
}

func (m *User) Init(collection *UserCollection) *User {
	m.UserCollection = collection //for UserModel itself
	m.BaseModel = mongo.NewModel(collection, m)
	return m
}

func (m *User) GetBaseModel() *mongo.BaseModel {
	return m.BaseModel
}

func (m *User) SetBaseModel(bm *mongo.BaseModel) {
	m.BaseModel = bm
}

func (m *User) ClearModel() {
	m.TelegramID = 0
	m.Username = ""
	m.FirstName = ""
	m.LastName = ""
	m.Status = ""
	m.LastAdditionAt = time.Time{}
}

func (m *User) GetContent() bson.M {
	return bson.M{
		"telegram_id":      m.TelegramID,
		"username":         m.Username,
		"firstname":        m.FirstName,
		"lastname":         m.LastName,
		"status":           m.Status,
		"last_addition_at": m.LastAdditionAt,
	}
}

func (m *User) SetContent(theMap bson.M) {
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

	if LastAdditionAtI, okM := theMap["last_addition_at"]; okM {
		if LastAdditionAt, okC := LastAdditionAtI.(time.Time); okC {
			m.LastAdditionAt = LastAdditionAt
		}
	}
}

func (m *User) UpdateTelegramId() (err *eepkg.ExtendedError) {
	return m.UserCollection.UpdateTelegramId(m)
}

func (m *User) UpdateOneOrInsertTelegramId() (isUpdated bool, err *eepkg.ExtendedError) {
	return m.UserCollection.UpdateOneOrInsertTelegramId(m)
}

type UserCollection struct {
	*mongo.BaseCollection
}

func (c *UserCollection) Init(connection *mongo.Connection) {
	c.BaseCollection = mongo.NewCollection(connection /*c,*/, "mazimotaBot", "users", reflect.TypeOf(&User{}))
}

func (c *UserCollection) GetIndexes() []mgo.Index {
	return []mgo.Index{{
		Key:        []string{"telegram_id"},
		Background: false,
		DropDups:   true,
		Unique:     true,
	}}
}

func (c *UserCollection) EnsureIndexes() *eepkg.ExtendedError {
	return c.BaseCollection.EnsureIndexes(c.GetIndexes())
}

func (c *UserCollection) GetBaseCollection() *mongo.BaseCollection {
	return c.BaseCollection
}

func (c *UserCollection) FindOne(query bson.M) (obj *User, err *eepkg.ExtendedError) {
	obj = &User{}
	obj.Init(c)
	err = c.BaseCollection.FindOneModel(query, obj)
	return
}

func (c *UserCollection) FindOneByTelegramID(telegramID int) (obj *User, err *eepkg.ExtendedError) {
	return c.FindOne(bson.M{
		"telegram_id": telegramID,
	})
}

func (c *UserCollection) FindOneByUsername(username string) (obj *User, err *eepkg.ExtendedError) {
	return c.FindOne(bson.M{
		"username": username,
	})
}

func (c *UserCollection) PipeOne(pipeline interface{}) (obj *User, err *eepkg.ExtendedError) {
	obj = &User{}
	obj.Init(c)
	err = c.BaseCollection.PipeOneModel(pipeline, obj)
	return
}

func (c *UserCollection) Insert(values ...*User) *eepkg.ExtendedError {
	models := c.toModels(values)
	return c.BaseCollection.InsertModel(models...)
}

func (c *UserCollection) UpdateTelegramId(value *User) (err *eepkg.ExtendedError) {
	return c.BaseCollection.UpdateModel(bson.M{
		"telegram_id": value.TelegramID,
	}, value)
}

func (c *UserCollection) InsertOneOrUpdateTelegramId(value *User) (isUpdated bool, err *eepkg.ExtendedError) {
	return c.BaseCollection.InsertOneOrUpdateModel(bson.M{
		"telegram_id": value.TelegramID,
	}, value)
}

func (c *UserCollection) UpdateOneOrInsertTelegramId(value *User) (isUpdated bool, err *eepkg.ExtendedError) {
	return c.BaseCollection.UpdateOneOrInsertModel(bson.M{
		"telegram_id": value.TelegramID,
	}, value)
}

func (c *UserCollection) RemoveByTelegramID(telegramID int) *eepkg.ExtendedError {
	return c.RemoveInterface(bson.M{
		"telegram_id": telegramID,
	})
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
