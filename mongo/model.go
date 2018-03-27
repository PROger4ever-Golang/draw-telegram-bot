package mongo

import (
	"time"

	"github.com/PROger4ever/draw-telegram-bot/error"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var ModelIDNotSetError = eepkg.New(false, false, "Model's ID isn't set")

type Model interface {
	GetBaseModel() *BaseModel
	SetBaseModel(bm *BaseModel)

	GetContent() bson.M //reflection and Unmarshal can be used instead
	ClearModel()
	SetContent(theMap bson.M)
}

type BaseModel struct {
	Collection Collection
	Model      Model

	ID        bson.ObjectId `bson:"_id,omitempty"`
	CreatedAt time.Time     `bson:"created_at,omitempty"`
	UpdatedAt time.Time     `bson:"updated_at,omitempty"`
	DeletedAt time.Time     `bson:"deleted_at,omitempty"`
}

func (m *BaseModel) Init(collection Collection, model Model) *BaseModel {
	m.Collection = collection
	m.Model = model
	return m
}

func (m *BaseModel) InitializeId() *BaseModel {
	if !m.ID.Valid() {
		m.ID = bson.NewObjectId()
	}
	return m
}

func (m *BaseModel) InitializeCommons() *BaseModel {
	t := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = t
	}
	m.UpdatedAt = t
	return m
}

func (m *BaseModel) UpdateDate() *BaseModel {
	t := time.Now()
	m.UpdatedAt = t
	return m
}

func (m *BaseModel) GetContent() (theMap bson.M) {
	theMap = m.Model.GetContent() //unsorted map :(
	if m.ID.Valid() {
		theMap["_id"] = m.ID
	}
	if !m.CreatedAt.IsZero() {
		theMap["created_at"] = m.CreatedAt
	}
	if !m.UpdatedAt.IsZero() {
		theMap["updated_at"] = m.UpdatedAt
	}
	if !m.DeletedAt.IsZero() {
		theMap["deleted_at"] = m.DeletedAt
	}
	return theMap
}

func (m *BaseModel) GetUpdateMap() (theMap bson.M) {
	theMap = m.Model.GetContent()
	if !m.UpdatedAt.IsZero() {
		theMap["updated_at"] = m.UpdatedAt
	}
	if !m.DeletedAt.IsZero() { //TODO: how to undelete Model, if omitempty set?
		theMap["deleted_at"] = m.DeletedAt
	}
	return theMap
}
func (m *BaseModel) SetContent(theMap bson.M) *BaseModel {
	m.Model.ClearModel()
	m.Model.SetContent(theMap)

	if idI, okM := theMap["_id"]; okM {
		if id, okC := idI.(bson.ObjectId); okC {
			m.ID = id
		}
	}

	if createdAtI, okM := theMap["created_at"]; okM {
		if createdAt, okC := createdAtI.(time.Time); okC {
			m.CreatedAt = createdAt
		}
	}

	if updatedAtI, okM := theMap["updated_at"]; okM {
		if updatedAt, okC := updatedAtI.(time.Time); okC {
			m.UpdatedAt = updatedAt
		}
	}

	if deletedAtI, okM := theMap["deleted_at"]; okM {
		if deletedAt, okC := deletedAtI.(time.Time); okC {
			m.CreatedAt = deletedAt
		}
	}
	return m
}

func (m *BaseModel) Upsert(query bson.M) (info *mgo.ChangeInfo, err *eepkg.ExtendedError) {
	theMap := m.InitializeCommons().GetContent()
	return m.Collection.GetBaseCollection().UpsertInterface(query, theMap)
}

func (m *BaseModel) UpsertId() (info *mgo.ChangeInfo, err *eepkg.ExtendedError) {
	theMap := m.InitializeId().InitializeCommons().GetContent()
	return m.Collection.GetBaseCollection().UpsertIdInterface(m.ID, theMap)
}

func (m *BaseModel) UpdateOneOrInsertModel(query bson.M) (isUpdated bool, err *eepkg.ExtendedError) {
	return m.Collection.GetBaseCollection().UpdateOneOrInsertModel(query, m.Model)
}

func (m *BaseModel) RemoveId() (err *eepkg.ExtendedError) {
	if !m.ID.Valid() {
		return ModelIDNotSetError
	}
	return m.Collection.GetBaseCollection().RemoveIdInterface(m.ID)
}

func NewModel(collection Collection, model Model) (m *BaseModel) {
	m = &BaseModel{}
	return m.Init(collection, model)
}
