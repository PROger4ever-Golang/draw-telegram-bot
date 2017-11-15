package mongo

import (
	"fmt"
	"reflect"
	"time"

	"bitbucket.org/proger4ever/draw-telegram-bot/error"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const cantEnsureIndexes = "Can't ensure indexes"
const cantQueryDB = "Ошибка при операции с БД"

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
	EnsureIndexes() *eepkg.ExtendedError

	GetBaseCollection() *BaseCollection
}

type BaseCollection struct {
	Connection *Connection
	//Collection Collection //it's a pointer actually
	DbName string
	Name   string
	Type   reflect.Type
}

func (c *BaseCollection) Init(connection *Connection /*Collection Collection,*/, dbName string, name string, theType reflect.Type) *BaseCollection {
	c.Connection = connection
	//c.Collection = Collection
	c.DbName = dbName
	c.Name = name
	c.Type = theType
	return c
}

func (c *BaseCollection) EnsureIndexes(indexes []mgo.Index) (err *eepkg.ExtendedError) {
	var errStd error
	for i := 0; i < len(indexes) && err == nil; i++ {
		errStd = c.Connection.DB(c.DbName).C(c.Name).EnsureIndex(indexes[i])
	}
	return eepkg.Wrap(errStd, false, true, cantEnsureIndexes)
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

func (c *BaseCollection) FindOneInterface(query bson.M, value interface{}) (err *eepkg.ExtendedError) {
	errStd := c.Connection.DB(c.DbName).C(c.Name).Find(query).One(value)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) FindOneModel(query bson.M, model Model) (err *eepkg.ExtendedError) {
	dataMap := bson.M{}
	errStd := c.Connection.DB(c.DbName).C(c.Name).Find(query).One(dataMap)
	model.GetBaseModel().SetContent(dataMap)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) CountInterface(query bson.M) (n int, err *eepkg.ExtendedError) {
	n, errStd := c.Connection.DB(c.DbName).C(c.Name).Find(query).Count()
	return n, eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) InsertInterface(values ...interface{}) (err *eepkg.ExtendedError) {
	errStd := c.Connection.DB(c.DbName).C(c.Name).Insert(values)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) InsertModel(models ...Model) (err *eepkg.ExtendedError) {
	c.CheckTypes(models)
	maps := c.getModelMaps(models)
	errStd := c.Connection.DB(c.DbName).C(c.Name).Insert(maps...)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) InsertOneOrUpdateModel(query bson.M, model Model) (isUpdated bool, err *eepkg.ExtendedError) {
	c.CheckType(model)

	bm := model.GetBaseModel()
	bm.InitializeId().InitializeCommons()
	theMap := bm.GetContent()

	errStd := c.Connection.DB(c.DbName).C(c.Name).Insert(theMap)
	isUpdated = mgo.IsDup(errStd)
	if isUpdated {
		bm.ID = bson.ObjectId("")
		bm.CreatedAt = time.Time{}

		theMap = bm.GetContent()
		newMap := bson.M{}
		_, errStd := c.Connection.DB(c.DbName).C(c.Name).Find(query).Apply(mgo.Change{
			Update: bson.M{
				"$set": theMap,
			},
			ReturnNew: true,
		}, newMap)
		if errStd != nil {
			return false, eepkg.Wrap(errStd, false, true, cantQueryDB)
		}
		bm.SetContent(newMap)
	}
	return isUpdated, eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) UpdateOneOrInsertModel(query bson.M, model Model) (isUpdated bool, err *eepkg.ExtendedError) {
	c.CheckType(model)

	bm := model.GetBaseModel()
	bm.UpdateDate()
	theMap := bm.GetUpdateMap()
	newMap := bson.M{}
	info, errStd := c.Connection.DB(c.DbName).C(c.Name).Find(query).Apply(mgo.Change{
		Update: bson.M{
			"$set": theMap,
		},
		ReturnNew: true,
	}, newMap)
	if errStd != nil {
		return false, eepkg.Wrap(errStd, false, true, cantQueryDB)
	}
	if info.Matched > 0 {
		bm.SetContent(newMap)
		return true, nil
	}

	bm.InitializeId().InitializeCommons()
	theMap = bm.GetContent()
	errStd = c.Connection.DB(c.DbName).C(c.Name).Insert(theMap)
	return false, eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) UpdateInterface(query bson.M, value interface{}) (err *eepkg.ExtendedError) {
	errStd := c.Connection.DB(c.DbName).C(c.Name).Update(query, value)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) UpdateModel(query bson.M, model Model) (err *eepkg.ExtendedError) {
	c.CheckType(model)
	theMap := c.getModelMap(model)
	errStd := c.Connection.DB(c.DbName).C(c.Name).Update(query, theMap)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) UpdateIdInterface(id interface{}, value interface{}) (err *eepkg.ExtendedError) {
	errStd := c.Connection.DB(c.DbName).C(c.Name).UpdateId(id, value)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) UpdateIdModel(id interface{}, model Model) (err *eepkg.ExtendedError) {
	c.CheckType(model)
	theMap := c.getModelMap(model)
	errStd := c.Connection.DB(c.DbName).C(c.Name).UpdateId(id, theMap)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) UpsertInterface(query bson.M, value interface{}) (info *mgo.ChangeInfo, err *eepkg.ExtendedError) {
	info, errStd := c.Connection.DB(c.DbName).C(c.Name).Upsert(query, value)
	return info, eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) UpsertModel(query bson.M, model Model) (info *mgo.ChangeInfo, err *eepkg.ExtendedError) {
	c.CheckType(model)
	theMap := c.getModelMap(model)
	info, errStd := c.Connection.DB(c.DbName).C(c.Name).Upsert(query, theMap)
	return info, eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) UpsertIdInterface(id interface{}, value interface{}) (info *mgo.ChangeInfo, err *eepkg.ExtendedError) {
	info, errStd := c.Connection.DB(c.DbName).C(c.Name).UpsertId(id, value)
	return info, eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) UpsertIdModel(id interface{}, model Model) (info *mgo.ChangeInfo, err *eepkg.ExtendedError) {
	c.CheckType(model)
	theMap := c.getModelMap(model)
	info, errStd := c.Connection.DB(c.DbName).C(c.Name).UpsertId(id, theMap)
	return info, eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) RemoveInterface(query bson.M) (err *eepkg.ExtendedError) {
	errStd := c.Connection.DB(c.DbName).C(c.Name).Remove(query)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) RemoveIdInterface(id bson.ObjectId) (err *eepkg.ExtendedError) {
	errStd := c.Connection.DB(c.DbName).C(c.Name).RemoveId(id)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) RemoveAllInterface(query bson.M) (err *eepkg.ExtendedError) {
	_, errStd := c.Connection.DB(c.DbName).C(c.Name).RemoveAll(query)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}

func (c *BaseCollection) PipeInterface(pipeline interface{}) *mgo.Pipe {
	return c.Connection.DB(c.DbName).C(c.Name).Pipe(pipeline)
}

func (c *BaseCollection) PipeOneModel(pipeline interface{}, model Model) (err *eepkg.ExtendedError) {
	dataMap := bson.M{}
	errStd := c.Connection.DB(c.DbName).C(c.Name).Pipe(pipeline).One(dataMap)
	model.GetBaseModel().SetContent(dataMap)
	return eepkg.Wrap(errStd, false, true, cantQueryDB)
}
func (c *BaseCollection) getModelMap(model Model) bson.M {
	bm := model.GetBaseModel()
	bm.InitializeId().InitializeCommons()
	return bm.GetContent()
}

func (c *BaseCollection) getModelMaps(models []Model) (maps []interface{}) {
	maps = make([]interface{}, 0, len(models))
	for _, m := range models {
		maps = append(maps, c.getModelMap(m))
	}
	return
}

func NewCollection(connection *Connection /*Collection Collection,*/, dbName string, name string, theType reflect.Type) (c *BaseCollection) {
	c = &BaseCollection{}
	return c.Init(connection /*Collection,*/, dbName, name, theType)
}

func NewCollectionDefault( /*Collection Collection,*/ dbName string, name string, theType reflect.Type) (c *BaseCollection) {
	c = &BaseCollection{}
	return c.Init(DefaultConnection /*Collection,*/, dbName, name, theType)
}
