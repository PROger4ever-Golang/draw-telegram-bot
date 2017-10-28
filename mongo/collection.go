package mongo

import (
	"fmt"
	"reflect"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type TypeMismatch struct {
	ActualType   reflect.Type
	ExpectedType reflect.Type
}

func (e *TypeMismatch) Error() string {
	return fmt.Sprintf("Expected type: %s, but got: %s", e.ExpectedType.Name(), e.ActualType.Name())
}

type Collection interface {
	Init(connection *Connection)
	GetIndexes() []mgo.Index
	EnsureIndexes() error
}

type BaseCollection struct {
	Connection *Connection
	Collection Collection //it's a pointer actually
	DbName     string
	Name       string
	Type       reflect.Type
}

func (c *BaseCollection) Init(connection *Connection, collection Collection, dbName string, name string, theType reflect.Type) *BaseCollection {
	c.Connection = connection
	c.Collection = collection
	c.DbName = dbName
	c.Name = name
	c.Type = theType
	return c
}

func (c *BaseCollection) EnsureIndexes(indexes []mgo.Index) (err error) {
	for i := 0; i < len(indexes) && err == nil; i++ {
		err = c.Connection.DB(c.DbName).C(c.Name).EnsureIndex(indexes[i])
		//TODO: return detailed error?
	}
	return err
}

func (c *BaseCollection) CheckType(value interface{}) {
	valueType := reflect.TypeOf(value)
	if c.Type != valueType {
		panic(&TypeMismatch{
			ExpectedType: c.Type,
			ActualType:   valueType,
		})
	}
}

func (c *BaseCollection) CheckTypes(values ...interface{}) {
	for v := range values {
		c.CheckType(v)
	}
}

func (c *BaseCollection) FindOneUnsafe(query bson.M, value interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	err = c.Connection.DB(c.DbName).C(c.Name).Find(query).One(value)
	return
}

func (c *BaseCollection) InsertUnsafe(values ...interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	err = c.Connection.DB(c.DbName).C(c.Name).Insert(values)
	return
}

func (c *BaseCollection) InsertSafe(values ...interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	c.CheckTypes(values)
	err = c.Connection.DB(c.DbName).C(c.Name).Insert(values)
	return
}

func (c *BaseCollection) UpsertUnsafe(query bson.M, value interface{}) (info *mgo.ChangeInfo, err error) {
	//connection state can be checked and reestablished with timout
	return c.Connection.DB(c.DbName).C(c.Name).Upsert(query, value)
}

func (c *BaseCollection) UpsertSafe(query bson.M, value interface{}) (info *mgo.ChangeInfo, err error) {
	//connection state can be checked and reestablished with timout
	c.CheckType(value)
	return c.Connection.DB(c.DbName).C(c.Name).Upsert(query, value)
}

func (c *BaseCollection) UpsertIdUnsafe(id interface{}, value interface{}) (info *mgo.ChangeInfo, err error) {
	//connection state can be checked and reestablished with timout
	return c.Connection.DB(c.DbName).C(c.Name).UpsertId(id, value)
}

func (c *BaseCollection) UpsertIdSafe(id interface{}, value interface{}) (info *mgo.ChangeInfo, err error) {
	//connection state can be checked and reestablished with timout
	c.CheckType(value)
	return c.Connection.DB(c.DbName).C(c.Name).UpsertId(id, value)
}

func (c *BaseCollection) RemoveUnsafe(query bson.M) error {
	//connection state can be checked and reestablished with timout
	return c.Connection.DB(c.DbName).C(c.Name).Remove(query)
}

func (c *BaseCollection) RemoveAllUnsafe(query bson.M) (err error) {
	//connection state can be checked and reestablished with timout
	_, err = c.Connection.DB(c.DbName).C(c.Name).RemoveAll(query)
	return
}

func NewCollection(connection *Connection, collection Collection, dbName string, name string, theType reflect.Type) (c *BaseCollection) {
	c = &BaseCollection{}
	return c.Init(connection, collection, dbName, name, theType)
}

func NewCollectionDefault(collection Collection, dbName string, name string, theType reflect.Type) (c *BaseCollection) {
	c = &BaseCollection{}
	return c.Init(DefaultConnection, collection, dbName, name, theType)
}
