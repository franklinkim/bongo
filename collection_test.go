package bongo

import (
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"testing"
)

type noHookDocument struct {
	DocumentBase `bson:",inline"`
	Name         string
}

type hookedDocument struct {
	DocumentBase    `bson:",inline"`
	RanBeforeSave   bool
	RanAfterSave    bool
	RanBeforeDelete bool
	RanAfterDelete  bool
	RanAfterFind    bool
}

func (h *hookedDocument) BeforeSave(c *Collection) error {
	h.RanBeforeSave = true
	return nil
}

func (h *hookedDocument) AfterSave(c *Collection) error {
	h.RanAfterSave = true
	return nil
}

func (h *hookedDocument) BeforeDelete(c *Collection) error {
	h.RanBeforeDelete = true
	return nil
}

func (h *hookedDocument) AfterDelete(c *Collection) error {
	h.RanAfterDelete = true
	return nil
}

func (h *hookedDocument) AfterFind(c *Collection) error {
	h.RanAfterFind = true
	return nil
}

type validatedDocument struct {
	DocumentBase `bson:",inline"`
	Name         string
}

func (v *validatedDocument) Validate(c *Collection) []string {
	return []string{"test validation error"}
}

func TestCollection(t *testing.T) {

	conn := getConnection()
	defer conn.Session.Close()

	Convey("Saving", t, func() {
		Convey("should be able to save a document with no hooks, update id, and use new tracker", func() {

			doc := &noHookDocument{}
			doc.Name = "foo"
			So(doc.IsNew(), ShouldEqual, true)

			err := conn.Collection("tests").Save(doc)
			So(err, ShouldEqual, nil)
			So(doc.ID.Valid(), ShouldEqual, true)
			So(doc.IsNew(), ShouldEqual, false)
		})

		Convey("should be able to save a document with save hooks", func() {
			doc := &hookedDocument{}

			err := conn.Collection("tests").Save(doc)

			So(err, ShouldEqual, nil)
			So(doc.RanBeforeSave, ShouldEqual, true)
			So(doc.RanAfterSave, ShouldEqual, true)
		})

		Convey("should return a validation error if the validate method has things in the return value", func() {
			doc := &validatedDocument{}
			err := conn.Collection("tests").Save(doc)

			v, ok := err.(*ValidationError)
			So(ok, ShouldEqual, true)
			So(v.Errors[0], ShouldEqual, "test validation error")
		})

		Convey("should be able to save an existing document", func() {
			doc := &noHookDocument{}
			doc.Name = "foo"
			So(doc.IsNew(), ShouldEqual, true)

			err := conn.Collection("tests").Save(doc)
			So(err, ShouldEqual, nil)
			So(doc.ID.Valid(), ShouldEqual, true)
			So(doc.IsNew(), ShouldEqual, false)

			err = conn.Collection("tests").Save(doc)

			So(err, ShouldEqual, nil)
			count, err := conn.Collection("tests").Collection().Count()
			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 1)
		})

		Convey("should set created and modified dates", func() {

			doc := &noHookDocument{}
			doc.Name = "foo"

			err := conn.Collection("tests").Save(doc)
			So(err, ShouldEqual, nil)
			So(doc.Created.UnixNano(), ShouldEqual, doc.Modified.UnixNano())

			err = conn.Collection("tests").Save(doc)
			So(err, ShouldEqual, nil)
			So(doc.Modified.UnixNano(), ShouldBeGreaterThan, doc.Created.UnixNano())
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})

	Convey("FindByID", t, func() {
		doc := &noHookDocument{}
		err := conn.Collection("tests").Save(doc)
		So(err, ShouldEqual, nil)

		Convey("should find a doc by id", func() {
			newDoc := &noHookDocument{}
			err := conn.Collection("tests").FindByID(doc.GetID(), newDoc)
			So(err, ShouldEqual, nil)
			So(newDoc.ID.Hex(), ShouldEqual, doc.ID.Hex())
		})

		Convey("should find a doc by id and run afterFind", func() {
			newDoc := &hookedDocument{}
			err := conn.Collection("tests").FindByID(doc.GetID(), newDoc)
			So(err, ShouldEqual, nil)
			So(newDoc.ID.Hex(), ShouldEqual, doc.ID.Hex())
			So(newDoc.RanAfterFind, ShouldEqual, true)
		})

		Convey("should return a document not found error if doc not found", func() {

			err := conn.Collection("tests").FindByID(bson.NewObjectId(), doc)
			_, ok := err.(*DocumentNotFoundError)
			So(ok, ShouldEqual, true)
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})

	Convey("FindOne", t, func() {
		doc := &noHookDocument{}
		doc.Name = "foo"
		err := conn.Collection("tests").Save(doc)
		So(err, ShouldEqual, nil)

		Convey("should find one with query", func() {
			newDoc := &noHookDocument{}
			err := conn.Collection("tests").FindOne(bson.M{
				"name": "foo",
			}, newDoc)
			So(err, ShouldEqual, nil)
			So(newDoc.ID.Hex(), ShouldEqual, doc.ID.Hex())
		})

		Convey("should find one with query and run afterFind", func() {
			newDoc := &hookedDocument{}
			err := conn.Collection("tests").FindOne(bson.M{
				"name": "foo",
			}, newDoc)
			So(err, ShouldEqual, nil)
			So(newDoc.ID.Hex(), ShouldEqual, doc.ID.Hex())
			So(newDoc.RanAfterFind, ShouldEqual, true)
		})

		Reset(func() {
			conn.Session.DB("bongotest").DropDatabase()
		})
	})

	Convey("Delete", t, func() {
		Convey("should be able delete a document", func() {
			doc := &noHookDocument{}

			err := conn.Collection("tests").Save(doc)
			So(err, ShouldEqual, nil)

			err = conn.Collection("tests").Delete(doc)
			So(err, ShouldEqual, nil)

			count, err := conn.Collection("tests").Collection().Count()

			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 0)
		})

		Convey("should be able delete a document and run hooks", func() {
			doc := &hookedDocument{}

			err := conn.Collection("tests").Save(doc)
			So(err, ShouldEqual, nil)

			err = conn.Collection("tests").Delete(doc)
			So(err, ShouldEqual, nil)

			count, err := conn.Collection("tests").Collection().Count()

			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 0)

			So(doc.RanBeforeDelete, ShouldEqual, true)
			So(doc.RanAfterDelete, ShouldEqual, true)
		})
	})
}
