package couchdb

import (
	"fmt"
	"io"
	"net/url"
)

// Does a "raw" GET, returning an io.Reader that can be used to parse the returned data yourself.
func (db *CouchDB) GetRaw(path string) (io.Reader, error) {
	escapedPath := url.QueryEscape(path)
	req, err := db.request("GET", escapedPath, nil)
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

// Accepts a struct or a map[string]something to fill with the doc's data, and a docid path relative to the database, returns error status
func (db *CouchDB) GetDocument(doc interface{}, path string) error {
	escapedPath := url.QueryEscape(path)
	return db.get(doc, escapedPath)
}

// As GetDocument, but also accepts a rev. Will attempt to retrieve the document at that rev
func (db *CouchDB) GetDocumentAtRev(doc interface{}, path, rev string) error {
	escapedPath := url.QueryEscape(path)
	return db.get(doc, fmt.Sprintf("%s?rev=%s", escapedPath, rev))
}
