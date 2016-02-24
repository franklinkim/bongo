package bongo

import (
	"errors"
	"fmt"
	"labix.org/v2/mgo"
)

// Config struct
type Config struct {
	ConnectionString string
	Database         string
}

// Connection struct
type Connection struct {
	Config  *Config
	Session *mgo.Session
}

// Connect creates a new connection and run Connect()
func Connect(config *Config) (*Connection, error) {
	conn := &Connection{
		Config: config,
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
	session, err := mgo.Dial(m.Config.ConnectionString)

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
		Name:       name,
	}
}
