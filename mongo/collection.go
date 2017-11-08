package mongo

import (
	"fmt"
	"reflect"
	"time"

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

func (c *BaseCollection) CheckType(model Model) {
	valueType := reflect.TypeOf(model)
	if c.Type != valueType {
		panic(&TypeMismatch{
			ExpectedType: c.Type,
			ActualType:   valueType,
		})
	}
}

func (c *BaseCollection) CheckTypes(values []Model) {
	for _, v := range values {
		c.CheckType(v)
	}
}

func (c *BaseCollection) FindOneInterface(query bson.M, value interface{}) (err error) {
	err = c.Connection.DB(c.DbName).C(c.Name).Find(query).One(value)
	return
}

func (c *BaseCollection) FindOneModel(query bson.M, model Model) (err error) {
	dataMap := bson.M{}
	err = c.Connection.DB(c.DbName).C(c.Name).Find(query).One(dataMap)
	model.GetBaseModel().SetContent(dataMap)
	return
}

func (c *BaseCollection) InsertInterface(values ...interface{}) (err error) {
	err = c.Connection.DB(c.DbName).C(c.Name).Insert(values)
	return
}

func (c *BaseCollection) InsertModel(models ...Model) (err error) {
	c.CheckTypes(models)
	maps := c.getModelMaps(models)
	err = c.Connection.DB(c.DbName).C(c.Name).Insert(maps...)
	return
}

func (c *BaseCollection) InsertOneOrUpdateModel(query bson.M, model Model) (isUpdated bool, err error) {
	c.CheckType(model)

	bm := model.GetBaseModel()
	bm.InitializeId().InitializeCommons()
	theMap := bm.GetContentMap()

	err = c.Connection.DB(c.DbName).C(c.Name).Insert(theMap)
	isUpdated = mgo.IsDup(err)
	if isUpdated {
		bm.ID = bson.ObjectId("")
		bm.CreatedAt = time.Time{}

		theMap = bm.GetContentMap()
		newMap := bson.M{}
		_, err = c.Connection.DB(c.DbName).C(c.Name).Find(query).Apply(mgo.Change{
			Update: bson.M{
				"$set": theMap,
			},
			ReturnNew: true,
		}, newMap)
		if err != nil {
			return false, err
		}
		bm.SetContent(newMap)
	}
	return
}

func (c *BaseCollection) UpdateInterface(query bson.M, value interface{}) (err error) {
	return c.Connection.DB(c.DbName).C(c.Name).Update(query, value)
}

func (c *BaseCollection) UpdateModel(query bson.M, model Model) (err error) {
	c.CheckType(model)
	theMap := c.getModelMap(model)
	return c.Connection.DB(c.DbName).C(c.Name).Update(query, theMap)
}

func (c *BaseCollection) UpdateIdInterface(id interface{}, value interface{}) (err error) {
	return c.Connection.DB(c.DbName).C(c.Name).UpdateId(id, value)
}

func (c *BaseCollection) UpdateIdModel(id interface{}, model Model) (err error) {
	c.CheckType(model)
	theMap := c.getModelMap(model)
	return c.Connection.DB(c.DbName).C(c.Name).UpdateId(id, theMap)
}

func (c *BaseCollection) UpsertInterface(query bson.M, value interface{}) (info *mgo.ChangeInfo, err error) {
	return c.Connection.DB(c.DbName).C(c.Name).Upsert(query, value)
}

func (c *BaseCollection) UpsertModel(query bson.M, model Model) (info *mgo.ChangeInfo, err error) {
	c.CheckType(model)
	theMap := c.getModelMap(model)
	return c.Connection.DB(c.DbName).C(c.Name).Upsert(query, theMap)
}

func (c *BaseCollection) UpsertIdInterface(id interface{}, value interface{}) (info *mgo.ChangeInfo, err error) {
	return c.Connection.DB(c.DbName).C(c.Name).UpsertId(id, value)
}

func (c *BaseCollection) UpsertIdModel(id interface{}, model Model) (info *mgo.ChangeInfo, err error) {
	c.CheckType(model)
	theMap := c.getModelMap(model)
	return c.Connection.DB(c.DbName).C(c.Name).UpsertId(id, theMap)
}

func (c *BaseCollection) RemoveInterface(query bson.M) error {
	return c.Connection.DB(c.DbName).C(c.Name).Remove(query)
}

func (c *BaseCollection) RemoveIdInterface(id bson.ObjectId) error {
	return c.Connection.DB(c.DbName).C(c.Name).RemoveId(id)
}

func (c *BaseCollection) RemoveAllInterface(query bson.M) (err error) {
	_, err = c.Connection.DB(c.DbName).C(c.Name).RemoveAll(query)
	return
}

func (c *BaseCollection) PipeInterface(pipeline interface{}) *mgo.Pipe {
	return c.Connection.DB(c.DbName).C(c.Name).Pipe(pipeline)
}

func (c *BaseCollection) PipeOneModel(pipeline interface{}, model Model) (err error) {
	dataMap := bson.M{}
	err = c.Connection.DB(c.DbName).C(c.Name).Pipe(pipeline).One(dataMap)
	model.GetBaseModel().SetContent(dataMap)
	return
}
func (c *BaseCollection) getModelMap(model Model) bson.M {
	bm := model.GetBaseModel()
	bm.InitializeId().InitializeCommons()
	return bm.GetContentMap()
}

func (c *BaseCollection) getModelMaps(models []Model) (maps []interface{}) {
	maps = make([]interface{}, 0, len(models))
	for _, m := range models {
		maps = append(maps, c.getModelMap(m))
	}
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
