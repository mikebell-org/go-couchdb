package couchdb

import (
	"fmt"
)

type BulkCommitResponseRow struct {
	ID     string `json:"id"`
	Rev    string `json:"rev"`
	Error  string `json:"error"`
	Reason string `json:"reason"`
}

type BulkCommitResponse []BulkCommitResponseRow

type BulkCommit struct {
	AllOrNothing bool          `json:"all_or_nothing,omitempty"` // Not guaranteed on regular couchdb, not supported on cloudant. Generally avoid.
	NewEdits     *bool         `json:"new_edits"`                // For replication
	Docs         []interface{} `json:"docs"`
}

// BulkUpdate accepts a commit with a list of updates to make to the DB, and returns a list of responses showing the status of each commit
func (db *CouchDB) BulkUpdate(c *BulkCommit) (*BulkCommitResponse, error) {
	var s BulkCommitResponse
	r, errCh := jsonifyDoc(c)
	req, err := db.request("POST", "_bulk_docs", r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	code, cerr := couchDo(req, &s)
	if cerr != nil {
		return nil, cerr
	}
	if err = <-errCh; err != nil {
		return nil, err
	}
	if code != 201 {
		// FIXME Unexpected code. Do something?
	}
	return &s, nil
}

func (db *CouchDB) PutDocument(doc interface{}, path string) (*CouchSuccess, error) {
	var s CouchSuccess
	r, errCh := jsonifyDoc(doc)
	req, err := db.request("PUT", path, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	_, cerr := couchDo(req, &s)
	if cerr != nil {
		return nil, cerr
	}
	if err := <-errCh; err != nil {
		return nil, err
	}
	return &s, nil
}

func (db *CouchDB) PostDocument(doc interface{}) (*CouchSuccess, error) {
	var s CouchSuccess
	r, errCh := jsonifyDoc(doc)
	req, err := db.request("POST", "", r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	code, cerr := couchDo(req, &s)
	if cerr != nil {
		return nil, cerr
	}
	if err = <-errCh; err != nil {
		return nil, err
	}
	if code != 201 {
		// FIXME Unexpected code. Do something?
	}
	return &s, nil
}

func (db *CouchDB) DeleteDocument(path, rev string) (*CouchSuccess, error) {
	var s CouchSuccess
	req, err := db.request("DELETE", fmt.Sprintf("%s?rev=%s", path, rev), nil)
	if err != nil {
		return nil, err
	}
	code, cerr := couchDo(req, &s)
	if cerr != nil {
		return nil, cerr
	}
	if code != 200 {
		// FIXME Unexpected code. Do something?
	}
	return &s, nil
}
