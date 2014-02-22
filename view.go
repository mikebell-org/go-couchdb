package couchdb

import (
	"encoding/json"
	"fmt"
	"io"
)

type ViewResults struct {
	TotalRows int `json:"total_rows"`
	Offset    int `json:"offset"`
	UpdateSeq int `json:"update_seq"`
	Rows      []ViewRow
}

type ViewRow struct {
	ID    string          `json:"id"`
	Key   interface{}     `json:"key"`
	Value interface{}     `json:"value"`
	Doc   json.RawMessage `json:"doc"`
}

type postViewData struct {
	Keys []interface{} `json:"keys"`
}

type ViewArgs struct {
	Key            interface{}     `urlencode:"key"`
	Keys           []interface{}   `urlencode:"keys"`
	StartKey       interface{}     `urlencode:"startkey"`
	StartKey_DocID string          `urlencode:"startkey_docid"`
	EndKey         interface{}     `urlencode:"endkey"`
	EndKey_DocID   string          `urlencode:"endkey_docid"`
	Limit          uint            `urlencode:"limit"`
	Stale          UnescapedString `urlencode:"stale"` // Special string because we don't want to quote this one
	Descending     bool            `urlencode:"descending"`
	Skip           uint            `urlencode:"skip"`
	Group          bool            `urlencode:"group"`
	GroupLevel     uint            `urlencode:"group_level"`
	Reduce         *bool           `urlencode:"reduce"` // Because the default is true
	IncludeDocs    bool            `urlencode:"include_docs"`
	InclusiveEnd   *bool           `urlencode:"inclusive_end"` // Because the default is true
	UpdateSeq      bool            `urlencode:"update_seq"`
}

// Encodes a ViewArgs struct into a query string for a view
func (a *ViewArgs) Encode() (string, error) {
	return urlEncodeObject(*a)
}

type AllDocsResult struct {
	TotalRows int `json:"total_rows"`
	Offset    int `json:"offset"`
	UpdateSeq int `json:"update_seq"`
	Rows      []AllDocsRow
}

type AllDocsRow struct {
	ID    string          `json:"id"`
	Key   interface{}     `json:"key"`
	Value AllDocsValue    `json:"value"`
	Doc   json.RawMessage `json:"doc"`
}

type AllDocsValue struct {
	Rev string `json:"rev"`
}

// Perform a query against _all_docs, such as a bulk get
func (db *CouchDB) AllDocs(args ViewArgs) (results *AllDocsResult, err error) {
	results = new(AllDocsResult)
	return results, db.viewHelper("_all_docs", args, results)
}

// Perform a view query
func (db *CouchDB) View(design, view string, args ViewArgs) (results *ViewResults, err error) {
	results = new(ViewResults)
	return results, db.viewHelper(fmt.Sprintf("_design/%s/_view/%s", design, view), args, results)
}

func (db *CouchDB) viewHelper(path string, args ViewArgs, results interface{}) (err error) {
	var argstring string
	var body io.Reader
	var errCh <-chan error
	method := "GET"

	if args.Keys != nil { // Always POST for keys= views to keep things simple, anyone have a scenario where this doesn't work?
		method = "POST"
		body, errCh = jsonifyDoc(postViewData{Keys: args.Keys})
		args.Keys = nil
	}
	if argstring, err = args.Encode(); err != nil {
		return err
	}
	requestString := fmt.Sprintf("%s?%s", path, argstring)
	req, err := db.request(method, requestString, body)
	if err != nil {
		return err
	}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	if _, err = couchDo(req, results); err != nil {
		return err
	}
	if errCh == nil {
		return
	}
	if err, _ := <-errCh; err != nil {
		return err
	}
	return
}
