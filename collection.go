package bongo

import (
	"errors"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// BeforeSaveHook ...
type BeforeSaveHook interface {
	BeforeSave(*Collection) error
}

// AfterSaveHook ...
type AfterSaveHook interface {
	AfterSave(*Collection) error
}

// BeforeDeleteHook ...
type BeforeDeleteHook interface {
	BeforeDelete(*Collection) error
}

// AfterDeleteHook ...
type AfterDeleteHook interface {
	AfterDelete(*Collection) error
}

// AfterFindHook ...
type AfterFindHook interface {
	AfterFind(*Collection) error
}

// ValidateHook ...
type ValidateHook interface {
	Validate(*Collection) []error
}

// ValidationError ...
type ValidationError struct {
	Errors []error
}

// TimeTracker ...
type TimeTracker interface {
	SetCreated(time.Time)
	SetModified(time.Time)
}

// Document ...
type Document interface {
	GetID() bson.ObjectId
	SetID(bson.ObjectId)
}

// CascadingDocument ...
type CascadingDocument interface {
	GetCascade(*Collection) []*CascadeConfig
}

func (v *ValidationError) Error() string {
	errs := make([]string, len(v.Errors))

	for i, e := range v.Errors {
		errs[i] = e.Error()
	}
	return "Validation failed. (" + strings.Join(errs, ", ") + ")"
}

// Collection ...
type Collection struct {
	Name       string
	Connection *Connection
}

// NewTracker ...
type NewTracker interface {
	SetIsNew(bool)
	IsNew() bool
}

// DocumentNotFoundError ...
type DocumentNotFoundError struct{}

func (d DocumentNotFoundError) Error() string {
	return "Document not found"
}

// Collection ...
func (c *Collection) Collection() *mgo.Collection {
	return c.Connection.Session.DB(c.Connection.Config.Database).C(c.Name)
}

func (c *Collection) collectionOnSession(sess *mgo.Session) *mgo.Collection {
	return sess.DB(c.Connection.Config.Database).C(c.Name)
}

// PreSave ...
func (c *Collection) PreSave(doc Document) error {
	// Validate?
	if validator, ok := doc.(ValidateHook); ok {
		errs := validator.Validate(c)

		if len(errs) > 0 {
			return &ValidationError{errs}
		}
	}

	if hook, ok := doc.(BeforeSaveHook); ok {
		err := hook.BeforeSave(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// Save ...
func (c *Collection) Save(doc Document) error {
	var err error
	sess := c.Connection.Session.Clone()
	defer sess.Close()

	// Per mgo's recommendation, create a clone of the session so there is no blocking
	col := c.collectionOnSession(sess)

	err = c.PreSave(doc)
	if err != nil {
		return err
	}
	// If the model implements the NewTracker interface, we'll use that to determine newness. Otherwise always assume it's new

	isNew := true
	if newt, ok := doc.(NewTracker); ok {
		isNew = newt.IsNew()
	}

	// Add created/modified time. Also set on the model itself if it has those fields.
	now := time.Now()

	if tt, ok := doc.(TimeTracker); ok {
		if isNew {
			tt.SetCreated(now)
		}
		tt.SetModified(now)
	}

	go CascadeSave(c, doc)

	id := doc.GetID()

	if !isNew && !id.Valid() {
		return errors.New("New tracker says this document isn't new but there is no valid ID field")
	}

	if isNew && !id.Valid() {
		// Generate an ID
		id = bson.NewObjectId()
		doc.SetID(id)
	}

	_, err = col.UpsertId(id, doc)

	if err != nil {
		return err
	}

	if hook, ok := doc.(AfterSaveHook); ok {
		err = hook.AfterSave(c)
		if err != nil {
			return err
		}
	}

	// We saved it, no longer new
	if newt, ok := doc.(NewTracker); ok {
		newt.SetIsNew(false)
	}

	return nil
}

// FindByID ...
func (c *Collection) FindByID(id bson.ObjectId, doc interface{}) error {

	err := c.Collection().FindId(id).One(doc)

	// Handle errors coming from mgo - we want to convert it to a DocumentNotFoundError so people can figure out
	// what the error type is without looking at the text
	if err != nil {
		if err.Error() == "not found" {
			return &DocumentNotFoundError{}
		}
		return err
	}

	if hook, ok := doc.(AfterFindHook); ok {
		err = hook.AfterFind(c)
		if err != nil {
			return err
		}
	}

	// We retrieved it, so set new to false
	if newt, ok := doc.(NewTracker); ok {
		newt.SetIsNew(false)
	}
	return nil
}

// Find doesn't actually do any DB interaction, it just creates the result set so we can
// start looping through on the iterator
func (c *Collection) Find(query interface{}) *ResultSet {
	col := c.Collection()

	// Count for testing
	q := col.Find(query)

	resultset := new(ResultSet)

	resultset.Query = q
	resultset.Params = query
	resultset.Collection = c

	return resultset
}

// FindOne ...
func (c *Collection) FindOne(query interface{}, doc interface{}) error {

	// Now run a find
	results := c.Find(query)
	results.Query.Limit(1)

	hasNext := results.Next(doc)

	if !hasNext {
		// There could have been an error fetching the next one, which would set the Error property on the resultset
		if results.Error != nil {
			return results.Error
		}
		return &DocumentNotFoundError{}
	}

	if newt, ok := doc.(NewTracker); ok {
		newt.SetIsNew(false)
	}

	return nil
}

// Delete ...
func (c *Collection) Delete(doc Document) error {
	var err error
	// Create a new session per mgo's suggestion to avoid blocking
	sess := c.Connection.Session.Clone()
	defer sess.Close()
	col := c.collectionOnSession(sess)

	if hook, ok := doc.(BeforeDeleteHook); ok {
		err := hook.BeforeDelete(c)
		if err != nil {
			return err
		}
	}

	err = col.Remove(bson.M{"_id": doc.GetID()})

	if err != nil {
		return err
	}

	go CascadeDelete(c, doc)

	if hook, ok := doc.(AfterDeleteHook); ok {
		err = hook.AfterDelete(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// RawDelete convenience method which just delegates to mgo.
// Note that hooks are NOT run
func (c *Collection) RawDelete(query bson.M) (*mgo.ChangeInfo, error) {
	sess := c.Connection.Session.Clone()
	defer sess.Close()
	col := c.collectionOnSession(sess)
	return col.RemoveAll(query)
}

// RawDeleteOne convenience method which just delegates to mgo.
// Note that hooks are NOT run
func (c *Collection) RawDeleteOne(query bson.M) error {
	sess := c.Connection.Session.Clone()
	defer sess.Close()
	col := c.collectionOnSession(sess)
	return col.Remove(query)
}
