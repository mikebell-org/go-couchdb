package couchdb

import "time"

// A minimal document you can embed into your document structs
type BasicDocument struct {
	ID  string `json:"_id,omitempty"`
	Rev string `json:"_rev,omitempty"`
}

// A more opinionated document you can base your structs off of.
// This library will take care of updating the Created and Modified fields for you in conjunction with the CouchDocPreWrite method
type BasicDocumentWithMtime struct {
	BasicDocument
	Created  float64 // Time document was created, expressed as decimal seconds since UNIX epoch
	Modified float64 // Time document was last modified, expressed as decimal seconds since UNIX epoch
}

func floatTime(t time.Time) (r float64) {
	return float64(t.Unix()) + (float64(t.Nanosecond()) / 1000000000.0)
}

func (d *BasicDocumentWithMtime) CouchDocPreWrite() {
	now := floatTime(time.Now())
	if d.Created == 0 {
		d.Created = now
	}
	d.Modified = now
}

type DocumentWithPreWriteHook interface {
	CouchDocPreWrite()
}
