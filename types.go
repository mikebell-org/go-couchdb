package couchdb

import (
	"os"
	"fmt"
	"http"
	"json"
)

func regularToCouchError(err os.Error) (e *CouchError) {
	e = new(CouchError)
	e.Error = err.String()
	return e
}

func responseToCouchError(r *http.Response) (e *CouchError) {
	e = new(CouchError)
	e.ReturnCode = r.StatusCode
	e.URL = r.Request.URL.String()
	j := json.NewDecoder(r.Body)
	err := j.Decode(e)
	if err != nil {
		e.Error = err.String()
	}
	return e
}

type CouchError struct {
	ReturnCode int
	URL        string
	Error      string `json:"error"`
	Reason     string `json:"reason"`
}

func (c *CouchError) String() (errstring string) {
	if c.ReturnCode == 0 {
		return c.Error
	}
	return fmt.Sprintf("URL: %s, HTTP response: %d, Error: %s, Reason: %s", c.URL, c.ReturnCode, c.Error, c.Reason)
}

type CouchSuccess struct {
	// {"ok":true,"id":"bob","rev":"1-967a00dff5e02add41819138abb3284d"}
	OK  bool   `json:"ok"`
	ID  string `json:"id"`
	Rev string `json:"rev"`
}

type DocRev struct {
	ID string `json:"id"`
	Seq   int    `json:"seq"`
	Doc	map[string]interface{}	`json:"doc"`
}


