package bongo

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/mgo.v2"
)

// For test usage
func getConnection() *Connection {
	dialInfo, err := mgo.ParseURL("mongodb://localhost:27017/bongotest")
	if err != nil {
		panic(err)
	}
	conn, err := Connect(dialInfo)
	conn.Context.Set("foo", "bar")
	if err != nil {
		panic(err)
	}
	return conn
}

func TestConnect(t *testing.T) {
	Convey("should be able to connect to a database using a config", t, func() {
		dialInfo, err := mgo.ParseURL("mongodb://localhost:27017/bongotest")
		So(err, ShouldBeNil)

		conn, err := Connect(dialInfo)
		defer conn.Session.Close()
		So(err, ShouldBeNil)

		conn.Context.Set("foo", "bar")
		value := conn.Context.Get("foo")
		So(value, ShouldEqual, "bar")

		err = conn.Session.Ping()
		So(err, ShouldBeNil)
	})
}

func TestRetrieveCollection(t *testing.T) {
	Convey("should be able to retrieve a collection instance from a connection", t, func() {
		conn := getConnection()
		defer conn.Session.Close()
		col := conn.Collection("tests")

		So(col.Name, ShouldEqual, "tests")
		So(col.Connection, ShouldEqual, conn)
	})
}
