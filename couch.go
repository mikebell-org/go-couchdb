// A library for accessing couchdb (and cloudant) from Go
package couchdb

import (
	"fmt"
	"io"
	"net/http"
)

type CouchDB struct {
	Host     string
	Database string
	Username string
	Password string
}

func (db *CouchDB) createRequest(method, urlpath, querystring string, body io.Reader) (r *http.Request, err error) {
	if r, err = http.NewRequest(method, db.Host, body); err != nil {
		return
	}
	opaque := fmt.Sprintf("//%s/%s/%s", r.URL.Host, db.Database, urlpath)
	if querystring != "" {
		opaque = fmt.Sprintf("%s?%s", opaque, querystring)
	}
	r.URL.Path = ""
	r.URL.RawQuery = ""
	r.URL.Opaque = opaque
	if db.Username != "" {
		r.SetBasicAuth(db.Username, db.Password)
	}
	r.Header.Set("Accept", "application/json")
	return
}

func (db *CouchDB) get(doc interface{}, path, query string) error {
	req, err := db.createRequest("GET", path, query, nil)
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
