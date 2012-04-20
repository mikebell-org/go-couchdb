package couchdb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func regularToCouchError(err error) (e *CouchError) {
	e = new(CouchError)
	e.Err = err.Error()
	return e
}

func responseToCouchError(r *http.Response) (e *CouchError) {
	e = new(CouchError)
	e.ReturnCode = r.StatusCode
	e.URL = r.Request.URL.String()
	j := json.NewDecoder(r.Body)
	err := j.Decode(e)
	if err != nil {
		e.Err = err.Error()
	}
	return e
}

type CouchError struct {
	ReturnCode int
	URL        string
	Err        string `json:"error"`
	Reason     string `json:"reason"`
}

func (c *CouchError) Error() (errstring string) {
	if c.ReturnCode == 0 {
		return c.Err
	}
	return fmt.Sprintf("URL: %s, HTTP response: %d, Error: %s, Reason: %s", c.URL, c.ReturnCode, c.Err, c.Reason)
}

type CouchSuccess struct {
	// {"ok":true,"id":"bob","rev":"1-967a00dff5e02add41819138abb3284d"}
	OK  bool   `json:"ok"`
	ID  string `json:"id"`
	Rev string `json:"rev"`
}

type DocRev struct {
	ID  string                 `json:"id"`
	Seq int                    `json:"seq"`
	Doc map[string]interface{} `json:"doc"`
}

type ViewResults struct {
	TotalRows int `json:"total_rows"`
	Offset    int `json:"offset"`
	Rows      []ViewRow
}

type ViewRow struct {
	ID    string                 `json:"id"`
	Key   interface{}            `json:"key"`
	Value interface{}            `json:"value"`
	Doc   map[string]interface{} `json:"doc"`
}
