package mongo

import (
	"errors"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Model interface {
	GetBaseModel() *BaseModel
	SetBaseModel(bm *BaseModel)

	GetContentMap() bson.M //reflection and Unmarshal can be used instead
	CleanModel()
	SetContentFromMap(theMap bson.M)
}

type BaseModel struct {
	collection *BaseCollection
	model      Model

	ID        bson.ObjectId `bson:"_id,omitempty"`
	CreatedAt time.Time     `bson:"created_at,omitempty"`
	UpdatedAt time.Time     `bson:"updated_at,omitempty"`
	DeletedAt time.Time     `bson:"deleted_at,omitempty"`
}

func (m *BaseModel) Init(collection *BaseCollection, model Model) *BaseModel {
	m.collection = collection
	m.model = model
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

func (m *BaseModel) GetContentMap() (theMap bson.M) {
	theMap = m.model.GetContentMap() //unsorted map :(
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

func (m *BaseModel) SetContentFromMap(theMap bson.M) *BaseModel {
	m.model.CleanModel()
	m.model.SetContentFromMap(theMap)

	if idI, okM := theMap["_id"]; okM {
		if id, okC := idI.(bson.ObjectId); okC {
			m.ID = id
		}
	}

	if createdAtI, okM := theMap["created_date"]; okM {
		if createdAt, okC := createdAtI.(time.Time); okC {
			m.CreatedAt = createdAt
		}
	}

	if createdAtI, okM := theMap["created_date"]; okM {
		if createdAt, okC := createdAtI.(time.Time); okC {
			m.CreatedAt = createdAt
		}
	}

	if deletedAtI, okM := theMap["deleted_date"]; okM {
		if deletedAt, okC := deletedAtI.(time.Time); okC {
			m.CreatedAt = deletedAt
		}
	}
	return m
}

func (m *BaseModel) Upsert(query bson.M) (info *mgo.ChangeInfo, err error) {
	theMap := m.InitializeCommons().GetContentMap()
	return m.collection.UpsertInterface(query, theMap)
}

func (m *BaseModel) UpsertId() (info *mgo.ChangeInfo, err error) {
	theMap := m.InitializeId().InitializeCommons().GetContentMap()
	return m.collection.UpsertIdInterface(m.ID, theMap)
}

func (m *BaseModel) RemoveId() (err error) {
	if !m.ID.Valid() {
		return errors.New("Model's ID isn't set")
	}
	return m.collection.RemoveIdInterface(m.ID)
}

func NewModel(collection *BaseCollection, model Model) (m *BaseModel) {
	m = &BaseModel{}
	return m.Init(collection, model)
}
