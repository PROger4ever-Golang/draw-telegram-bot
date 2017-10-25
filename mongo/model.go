package mongo

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// type ModelId interface {
// 	GetId() bson.ObjectId
// }

type Model struct {
	collection *Collection
	value      interface{}

	ID bson.ObjectId `bson:"_id,omitempty"`
}

func (m *Model) Init(collection *Collection, value interface{}) *Model {
	m.collection = collection
	m.value = value
	return m
}

func (m *Model) Upsert(query *bson.M) (info *mgo.ChangeInfo, err error) {
	if !m.ID.Valid() { //Empty
		m.ID = bson.NewObjectId()
	}
	return m.collection.UpsertUnsafe(query, m.value)
}

func (m *Model) UpsertId() (info *mgo.ChangeInfo, err error) {
	if !m.ID.Valid() { //Empty
		m.ID = bson.NewObjectId()
	}
	return m.collection.UpsertIdUnsafe(m.ID, m.value)
}

func NewModel(collection *Collection, value interface{}) (m *Model) {
	m = &Model{}
	return m.Init(collection, value)
}
