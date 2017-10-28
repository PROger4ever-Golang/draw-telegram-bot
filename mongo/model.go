package mongo

import (
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Model interface {
	GetContentMap() bson.M //reflection can be used instead
}

type BaseModel struct {
	collection *BaseCollection
	model      Model

	ID        bson.ObjectId `bson:"_id,omitempty"`
	CreatedAt time.Time     `bson:"created_at,omitempty"`
	UpdatedAt time.Time     `bson:"updated_at,omitempty"`
	DeletedAt time.Time     `bson:"deleted_at,omitempty"`
}

func (m *BaseModel) Init(collection *BaseCollection, value Model) *BaseModel {
	m.collection = collection
	m.model = value
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

func (m *BaseModel) Upsert(query bson.M) (info *mgo.ChangeInfo, err error) {
	theMap := m.initializeCommons().GetContentMap()
	return m.collection.UpsertUnsafe(query, theMap)
}

func (m *BaseModel) UpsertId() (info *mgo.ChangeInfo, err error) {
	theMap := m.initializeId().initializeCommons().GetContentMap()
	return m.collection.UpsertIdUnsafe(m.ID, theMap)
}

func (m *BaseModel) initializeId() *BaseModel {
	if !m.ID.Valid() {
		m.ID = bson.NewObjectId()
	}
	return m
}

func (m *BaseModel) initializeCommons() *BaseModel {
	t := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = t
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = t
	}
	return m
}

func NewModel(collection *BaseCollection, value Model) (m *BaseModel) {
	m = &BaseModel{}
	return m.Init(collection, value)
}
