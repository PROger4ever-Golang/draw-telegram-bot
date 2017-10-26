package mongo

import (
	"fmt"

	"bitbucket.org/proger4ever/draw-telegram-bot/common"
	mgo "gopkg.in/mgo.v2"
)

var DefaultConnection *Connection

type Connection struct {
	Host    string
	Port    int
	Session *mgo.Session
}

func (c *Connection) Init(host string, port int) (*Connection, error) {
	var err error
	c.Session, err = mgo.Dial(fmt.Sprintf("%s:%d", host, port))
	common.PanicIfError(err, "opening connection to mongo")
	c.Session.SetMode(mgo.Monotonic, true)
	fmt.Println("MongoSession opened")
	return c, err
}

func (c *Connection) DB(dbName string) *mgo.Database {
	//connection state can be checked and reestablished with timout
	return c.Session.DB(dbName)
}

func (c *Connection) Close() {
	if c.Session == nil {
		return
	}
	c.Session.Close()
}

func NewConnection(host string, port int) (connection *Connection, err error) {
	connection = &Connection{}
	return connection.Init(host, port)
}

func InitDefaultConnection(host string, port int) (connection *Connection, err error) {
	DefaultConnection, err = NewConnection(host, port)
	return DefaultConnection, err
}
