// A library for accessing couchdb (and cloudant) from Go
package couchdb

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
)

type CouchDB struct {
	Host     string
	Database string
	Username string
	Password string
}

func (db *CouchDB) request(method, urlpath string, body io.Reader) (r *http.Request, err error) {
	clean_url := func(url string) string {
		if strings.HasPrefix(url, "http://") {
			return "http://" + path.Clean(url[7:])
		} else if strings.HasPrefix(url, "https://") {
			return "https://" + path.Clean(url[8:])
		} else {
			return path.Clean(url)
		}
		panic("Shouldn't reach this spot")
	}

	url := clean_url(fmt.Sprintf("%s/%s/%s", db.Host, db.Database, urlpath))
	r, err = http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	if db.Username != "" {
		r.SetBasicAuth(db.Username, db.Password)
	}
	return
}

func (db *CouchDB) get(doc interface{}, path string) error {
	req, err := db.request("GET", path, nil)
	if err != nil {
		return err
	}
	code, cerr := couchDo(req, doc)
	if cerr != nil {
		return cerr
	}
	if code != 200 {
		// FIXME Unexpected code. Do something?
	}
	return nil
}

// Does a "raw" GET, returning an io.Reader that can be used to parse the returned data yourself.
func (db *CouchDB) GetRaw(path string) (io.Reader, error) {
	req, err := db.request("GET", path, nil)
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
	return db.get(doc, path)
}
