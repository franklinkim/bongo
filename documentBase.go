package bongo

import (
	"labix.org/v2/mgo/bson"
	"time"
)

// DocumentBase ...
type DocumentBase struct {
	ID       bson.ObjectId `bson:"_id" json:"_id"`
	Created  time.Time     `bson:"_created" json:"_created"`
	Modified time.Time     `bson:"_modified" json:"_modified"`

	// We want this to default to false without any work. So this will be the opposite of isNew. We want it to be new unless set to existing
	exists bool
}

// SetIsNew satisfy the new tracker interface
func (d *DocumentBase) SetIsNew(isNew bool) {
	d.exists = !isNew
}

// IsNew ...
func (d *DocumentBase) IsNew() bool {
	return !d.exists
}

// GetID satisfy the document interface
func (d *DocumentBase) GetID() bson.ObjectId {
	return d.ID
}

// SetID ...
func (d *DocumentBase) SetID(id bson.ObjectId) {
	d.ID = id
}

// SetCreated ...
func (d *DocumentBase) SetCreated(t time.Time) {
	d.Created = t
}

// SetModified ...
func (d *DocumentBase) SetModified(t time.Time) {
	d.Modified = t
}
