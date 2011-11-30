package couchdb

import (
	"os"
	"fmt"
	"http"
	"json"
)

func Database(host, database string) (db *CouchDB, err *CouchError) {
	db = new(CouchDB)
	db.Host = host
	db.Database = database
	return db, nil
}

func CreateDatabase(host, database string) (*CouchDB, *CouchError) {
	var s CouchSuccess
	url := fmt.Sprintf("%s/%s", host, database)
	db, cerr := Database(host, database)
	if cerr != nil {
		return nil, cerr
	}
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return nil, regularToCouchError(err)
	}
	code, cerr := couchDo(req, &s)
	if cerr != nil {
		return nil, cerr
	}
	if code != 201 {
		// FIXME Unexpected code. Do something?
	}
	return db, nil
}

type CouchDB struct {
	Host     string
	Database string
}

func (db *CouchDB) get(doc interface{}, path string) *CouchError {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/%s", db.Host, db.Database, path), nil)
	if err != nil {
		return regularToCouchError(err)
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

func (db *CouchDB) Delete() *CouchError {
	url := fmt.Sprintf("%s/%s", db.Host, db.Database)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return regularToCouchError(err)
	}
	code, cerr := couchDo(req, nil)
	if cerr != nil {
		return cerr
	}
	if code != 200 {
		// FIXME Unexpected code. Do something?
	}
	return nil
}

func (db *CouchDB) GetDocument(doc interface{}, path string) *CouchError {
	return db.get(doc, path)
}

func (db *CouchDB) PutDocument(doc interface{}, path string) (*CouchSuccess, *CouchError) {
	var s CouchSuccess
	r, errCh := jsonifyDoc(doc)
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s/%s", db.Host, db.Database, path), r)
	if err != nil {
		return nil, regularToCouchError(err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, cerr := couchDo(req, &s)
	if cerr != nil {
		return nil, cerr
	}
	if err := <-errCh; err != nil {
		return nil, regularToCouchError(err)
	}
	return &s, nil
}

func (db *CouchDB) PostDocument(doc interface{}) (*CouchSuccess, *CouchError) {
	var s CouchSuccess
	r, errCh := jsonifyDoc(doc)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/", db.Host, db.Database), r)
	if err != nil {
		return nil, regularToCouchError(err)
	}
	req.Header.Set("Content-Type", "application/json")
	code, cerr := couchDo(req, &s)
	if cerr != nil {
		return nil, cerr
	}
	if err = <-errCh; err != nil {
		return nil, regularToCouchError(err)
	}
	if code != 201 {
		// FIXME Unexpected code. Do something?
	}
	return &s, nil
}

func (db *CouchDB) DeleteDocument(path, rev string) (*CouchSuccess, *CouchError) {
	var s CouchSuccess
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s/%s?rev=%s", db.Host, db.Database, path, rev), nil)
	if err != nil {
		return nil, regularToCouchError(err)
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

func (db *CouchDB) ContinuousChanges(c chan *DocRev, since int, filter string) *CouchError {
	defer close(c)
	var url string
	if filter == "" {
		url = fmt.Sprintf("%s/%s/_changes?feed=continuous&since=%d&heartbeat=30000", db.Host, db.Database, since)
	} else {
		url = fmt.Sprintf("%s/%s/_changes?feed=continuous&since=%d&filter=%s&heartbeat=30000", db.Host, db.Database, since, filter)
	}
	r, err := http.Get(url)
	if err != nil {
		return regularToCouchError(err)
	}
	if r.StatusCode != 200 {
		return responseToCouchError(r)
	}
	j := json.NewDecoder(r.Body)
	seq := since
	for {
		var r DocRev
		err := j.Decode(&r)
		if err != nil {
			return regularToCouchError(err)
		}
		if seq >= r.Seq {
			return regularToCouchError(os.NewError(fmt.Sprintf("Sequence number was %d, but latest line (%s) from couch has it at %d.", seq, r, r.Seq)))
		}
		c <- &r
	}
	return regularToCouchError(os.NewError("This should be impossible to reach, just putting it here to shut up go"))
}
