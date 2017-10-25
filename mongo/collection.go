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

type Collection struct {
	Connection *Connection
	DbName     string
	Name       string
	Type       reflect.Type
}

func (c *Collection) Init(connection *Connection, dbName string, name string, theType reflect.Type) *Collection {
	c.Connection = connection
	c.DbName = dbName
	c.Name = name
	c.Type = theType
	return c
}

func (c *Collection) CheckType(value interface{}) {
	valueType := reflect.TypeOf(value)
	if c.Type != valueType {
		panic(&TypeMismatch{
			ExpectedType: c.Type,
			ActualType:   valueType,
		})
	}
}

func (c *Collection) CheckTypes(values ...interface{}) {
	for v := range values {
		c.CheckType(v)
	}
}

func (c *Collection) FindOneUnsafe(query *bson.M, value interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	err = c.Connection.DB(c.DbName).C(c.Name).Find(query).One(value)
	return
}

func (c *Collection) InsertUnsafe(values ...interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	err = c.Connection.DB(c.DbName).C(c.Name).Insert(values)
	return
}

func (c *Collection) InsertSafe(values ...interface{}) (err error) {
	//connection state can be checked and reestablished with timout
	c.CheckTypes(values)
	err = c.Connection.DB(c.DbName).C(c.Name).Insert(values)
	return
}

func (c *Collection) UpsertUnsafe(query *bson.M, value interface{}) (info *mgo.ChangeInfo, err error) {
	//connection state can be checked and reestablished with timout
	return c.Connection.DB(c.DbName).C(c.Name).Upsert(query, value)
}

func (c *Collection) UpsertSafe(query *bson.M, value interface{}) (info *mgo.ChangeInfo, err error) {
	//connection state can be checked and reestablished with timout
	c.CheckType(value)
	return c.Connection.DB(c.DbName).C(c.Name).Upsert(query, value)
}

func (c *Collection) UpsertIdUnsafe(id interface{}, value interface{}) (info *mgo.ChangeInfo, err error) {
	//connection state can be checked and reestablished with timout
	return c.Connection.DB(c.DbName).C(c.Name).UpsertId(id, value)
}

func (c *Collection) UpsertIdSafe(id interface{}, value interface{}) (info *mgo.ChangeInfo, err error) {
	//connection state can be checked and reestablished with timout
	c.CheckType(value)
	return c.Connection.DB(c.DbName).C(c.Name).UpsertId(id, value)
}

func (c *Collection) RemoveUnsafe(query *bson.M) error {
	//connection state can be checked and reestablished with timout
	return c.Connection.DB(c.DbName).C(c.Name).Remove(query)
}

func (c *Collection) RemoveAllUnsafe(query *bson.M) (err error) {
	//connection state can be checked and reestablished with timout
	_, err = c.Connection.DB(c.DbName).C(c.Name).RemoveAll(query)
	return
}

func NewCollection(connection *Connection, dbName string, name string, theType reflect.Type) (c *Collection) {
	c = &Collection{}
	return c.Init(connection, dbName, name, theType)
}
