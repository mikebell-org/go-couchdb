package couchdb

import (
	"time"

	"github.com/mikebell-org/timestamp"
)

type Attachment struct {
	ContentType string `json:"content_type"`
	RevPos      uint64 `json:"revpos"`
	Digest      string `json:"digest"`
	Length      uint64 `json:"length"`
	Stub        bool   `json:"stub"`
}

// A minimal document you can embed into your document structs
type BasicDocument struct {
	ID          string                `json:"_id,omitempty"`
	Rev         string                `json:"_rev,omitempty"`
	Deleted     bool                  `json:"_deleted,omitempty"`
	Attachments map[string]Attachment `json:"_attachments,omitempty"`
}

// A more opinionated document you can base your structs off of.
// This library will take care of updating the Created and Modified fields for you in conjunction with the CouchDocPreWrite method
type BasicDocumentWithMtime struct {
	BasicDocument
	Created  timestamp.Timestamp // Time document was created, expressed as decimal seconds since UNIX epoch
	Modified timestamp.Timestamp // Time document was last modified, expressed as decimal seconds since UNIX epoch
}

func (d *BasicDocumentWithMtime) CouchDocPreWrite() {
	now := timestamp.New(time.Now())
	if d.Created == 0 {
		d.Created = now
	}
	d.Modified = now
}

type DocumentWithPreWriteHook interface {
	CouchDocPreWrite()
}
