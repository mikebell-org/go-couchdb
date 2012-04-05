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

type CouchInfo struct {
	Name               string `json:"db_name"`
	DocCount           int    `json:"doc_count"`
	DocDelCount        int    `json:"doc_del_count"`
	UpdateSeq          interface{}    `json:"update_seq"`
	PurgeSeq           int    `json:"purge_seq"`
	CompactRunning     bool   `json:"compact_running"`
	DiskSize           int    `json:"disk_size"`
	InstanceStartTime  string `json:"instance_start_time"`
	DiskFormatVersion  int    `json:"disk_format_version"`
	CommittedUpdateSeq int    `json:"committed_update_seq"`
}

type DocRev struct {
	ID  string        `json:"id"`
	Seq interface{}   `json:"seq"`
	Doc CouchDocument `json:"doc"`
}

type ViewResults struct {
	TotalRows int `json:"total_rows"`
	Offset    int `json:"offset"`
	UpdateSeq int `json:"update_seq"`
	Rows      []ViewRow
}

type ViewRow struct {
	ID    string        `json:"id"`
	Key   interface{}   `json:"key"`
	Value interface{}   `json:"value"`
	Doc   CouchDocument `json:"doc"`
}

type BulkCommitResponseRow struct {
	ID     string `json:"id"`
	Rev    string `json:"rev"`
	Error  string `json:"error"`
	Reason string `json:"reason"`
}

type BulkCommitResponse []BulkCommitResponseRow

type BulkCommit struct {
	AllOrNothing bool            `json:"all_or_nothing,omitempty"`
	Docs         []CouchDocument `json:"docs"`
}

type CouchDocument map[string]interface{}
