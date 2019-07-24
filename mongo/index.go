package mongo

import (
	"fmt"

	"github.com/PROger4ever-Golang/draw-telegram-bot/error"
	"gopkg.in/mgo.v2"
)

const connectionFailed = "Connection to MongoDB failed"

var DefaultConnection *Connection

type Connection struct {
	Host    string
	Port    int
	Session *mgo.Session
}

func (c *Connection) Init(host string, port int) (*Connection, *eepkg.ExtendedError) {
	var errStd error
	c.Session, errStd = mgo.Dial(fmt.Sprintf("%s:%d", host, port))
	if errStd != nil {
		return c, eepkg.Wrap(errStd, false, true, connectionFailed)
	}
	c.Session.SetMode(mgo.Monotonic, true)
	return c, nil
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

func NewConnection(host string, port int) (connection *Connection, err *eepkg.ExtendedError) {
	connection = &Connection{}
	return connection.Init(host, port)
}

func InitDefaultConnection(host string, port int) (connection *Connection, err *eepkg.ExtendedError) {
	DefaultConnection, err = NewConnection(host, port)
	return DefaultConnection, err
}
