package baseModel

import (
	"fmt"
	"reflect"

	"bitbucket.org/proger4ever/draw-telegram-bot/mongo"
	"gopkg.in/mgo.v2/bson"
)

type TypeMismatch struct {
	ActualType   reflect.Type
	ExpectedType reflect.Type
}

func (e *TypeMismatch) Error() string {
	return fmt.Sprintf("Expected type: %s, but got: %s", e.ExpectedType.Name(), e.ActualType.Name())
}

type BaseModel struct {
	Connection     *mongo.Connection
	DbName         string
	CollectionName string
	Type           reflect.Type
}

func (m *BaseModel) Init(connection *mongo.Connection, dbName string, collectionName string, theType reflect.Type) *BaseModel {
	m.Connection = connection
	m.DbName = dbName
	m.CollectionName = collectionName
	m.Type = theType
	return m
}

func (m *BaseModel) CheckType(value interface{}) {
	valueType := reflect.TypeOf(value)
	if m.Type != valueType {
		panic(&TypeMismatch{
			ExpectedType: m.Type,
			ActualType:   valueType,
		})
	}
}

func (m *BaseModel) CheckTypes(values ...interface{}) {
	for v := range values {
		m.CheckType(v)
	}
}

func (m *BaseModel) FindOneUnsafe(query *bson.M, value interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	err = m.Connection.DB(m.DbName).C(m.CollectionName).Find(query).One(value)
	return
}

func (m *BaseModel) InsertUnsafe(values ...interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	err = m.Connection.DB(m.DbName).C(m.CollectionName).Insert(values)
	return
}

func (m *BaseModel) InsertSafe(values ...interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	m.CheckTypes(values)
	err = m.Connection.DB(m.DbName).C(m.CollectionName).Insert(values)
	return
}

func (m *BaseModel) UpsertUnsafe(query *bson.M, value interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	_, err = m.Connection.DB(m.DbName).C(m.CollectionName).Upsert(query, value)
	return
}

func (m *BaseModel) UpsertSafe(query *bson.M, value interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	m.CheckType(value)
	_, err = m.Connection.DB(m.DbName).C(m.CollectionName).Upsert(query, value)
	return
}

func (m *BaseModel) RemoveUnsafe(query *bson.M) error {
	//connection state can be checked and reestablished with timout
	return m.Connection.DB(m.DbName).C(m.CollectionName).Remove(query)
}

func (m *BaseModel) RemoveAllUnsafe(query *bson.M) (err error) {
	//connection state can be checked and reestablished with timout
	_, err = m.Connection.DB(m.DbName).C(m.CollectionName).RemoveAll(query)
	return
}

func New(connection *mongo.Connection, dbName string, collectionName string, theType reflect.Type) (m *BaseModel) {
	m = &BaseModel{}
	m.Init(connection, dbName, collectionName, theType)
	return
}
