package bongo

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

// DocumentBase ...
type DocumentBase struct {
	ID         bson.ObjectId `bson:"_id,omitempty" json:"_id"`
	CreatedAt  time.Time     `bson:"createdAt" json:"createdAt"`
	ModifiedAt time.Time     `bson:"modifiedAt" json:"modifiedAt"`

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
	d.CreatedAt = t
}

// SetModified ...
func (d *DocumentBase) SetModified(t time.Time) {
	d.ModifiedAt = t
}
