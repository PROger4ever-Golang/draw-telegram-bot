package baseModel

import (
	"bitbucket.org/proger4ever/drawtelegrambot/mongo"
	"gopkg.in/mgo.v2/bson"
)

type BaseModel struct {
	Connection     *mongo.Connection
	DbName         string
	CollectionName string
}

func (m *BaseModel) Init(connection *mongo.Connection, dbName string, collectionName string) *BaseModel {
	m.Connection = connection
	m.DbName = dbName
	m.CollectionName = collectionName
	return m
}

func (m *BaseModel) FindOne(query bson.M, value interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	err = m.Connection.DB(m.DbName).C(m.CollectionName).Find(query).One(value)
	return
}

func (m *BaseModel) Upsert(query bson.M, value interface{}) (err error) {
	_, err = m.Connection.DB(m.DbName).C(m.CollectionName).Upsert(query, value)
	return
}

func New(connection *mongo.Connection, dbName string, collectionName string) (m *BaseModel) {
	m = &BaseModel{}
	m.Init(connection, dbName, collectionName)
	return
}
