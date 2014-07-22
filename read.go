package couchdb

import (
	"io"
)

func (db *CouchDB) getRaw(path, query string) (io.Reader, error) {
	req, err := db.createRequest("GET", path, query, nil)
	if err != nil {
		return nil, err
	}
	r, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if r.StatusCode >= 400 {
		return nil, responseToCouchError(r)
	}
	return r.Body, nil
}

// Does a "raw" GET, returning an io.Reader that can be used to parse the returned data yourself.
func (db *CouchDB) GetRaw(path, query string) (io.Reader, error) {
	return db.getRaw(escape_docid(path), query)
}

// Same as GetRaw, does a "raw" GET, returning an io.Reader that can be used to parse the returned data yourself.
func (db *CouchDB) GetAttachment(docid, attname, query string) (io.Reader, error) {
	return db.getRaw(escape_docid(docid)+"/"+escape_docid(attname), query)
}

// Accepts a struct or a map[string]something to fill with the doc's data, and a docid path relative to the database, returns error status
func (db *CouchDB) GetDocument(doc interface{}, docid string) error {
	return db.get(doc, escape_docid(docid), "")
}

// As GetDocument, but also accepts a rev. Will attempt to retrieve the document at that rev
func (db *CouchDB) GetDocumentAtRev(doc interface{}, docid, rev string) error {
	return db.get(doc, escape_docid(docid), rev)
}
