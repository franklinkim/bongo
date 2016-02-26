package bongo

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
)

// Connection struct
type Connection struct {
	DialInfo *mgo.DialInfo
	Session  *mgo.Session
	Context  *Context
}

// Connect creates a new connection and run Connect()
func Connect(info *mgo.DialInfo) (*Connection, error) {
	set := make(map[string]interface{})
	conn := &Connection{
		DialInfo: info,
		Context: &Context{
			set: set,
		},
	}
	err := conn.Connect()
	return conn, err
}

// Connect to the database using the provided config
func (m *Connection) Connect() (err error) {
	defer func() {
		if r := recover(); r != nil {
			// panic(r)
			// return
			if e, ok := r.(error); ok {
				err = e
			} else if e, ok := r.(string); ok {
				err = errors.New(e)
			} else {
				err = errors.New(fmt.Sprint(r))
			}

		}
	}()
	session, err := mgo.DialWithInfo(m.DialInfo)

	if err != nil {
		return err
	}

	m.Session = session

	m.Session.SetMode(mgo.Monotonic, true)
	return nil
}

// Collection ...
func (m *Connection) Collection(name string) *Collection {
	// Just create a new instance - it's cheap and only has name
	return &Collection{
		Connection: m,
		Context:    m.Context,
		Name:       name,
	}
}
