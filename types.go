package couchdb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func responseToCouchError(r *http.Response) error {
	e := new(CouchError)
	e.ReturnCode = r.StatusCode
	e.URL = r.Request.URL.String()
	j := json.NewDecoder(r.Body)
	err := j.Decode(e)
	if err != nil {
		e.Err = err.Error()
	}
	return e
}

// A type describing an error returned by couchdb per CouchDB Error Status (plus fields for an HTTP return code and URL)
type CouchError struct {
	ReturnCode int
	URL        string
	ID         string `json:"id"`
	Err        string `json:"error"`
	Reason     string `json:"reason"`
}

func (c *CouchError) Error() string {
	if c.ReturnCode == 0 {
		if c.Err == "" {
			panic("Internal error in couchdb library, c.ReturnCode == 0 and c.Err == \"\"")
		}
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
