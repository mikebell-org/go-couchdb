package couchdb

import (
	"io"
	"fmt"
	"url"
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

func (db *CouchDB) GetRaw(path string) (io.Reader, *CouchError) {
	url := fmt.Sprintf("%s/%s/%s", db.Host, db.Database, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, regularToCouchError(err)
	}
	r, err := client.Do(req)
	if err != nil {
		return nil, regularToCouchError(err)
	}
	if r.StatusCode >= 400 {
		return nil, responseToCouchError(r)
	}
	return r.Body, nil
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

func (db *CouchDB) View(design, view string, args url.Values) (results *ViewResults, cerr *CouchError) {
	results = new(ViewResults)
	cerr = db.GetDocument(results, fmt.Sprintf("_design/%s/_view/%s?%s", design, view, args.Encode()))
	if cerr != nil {
		return nil, cerr
	}
	return
}

func (db *CouchDB) ContinuousChanges(args url.Values) (chan *DocRev, *CouchError) {
	c := make(chan *DocRev)
	args.Set("feed", "continuous")
	url := fmt.Sprintf("%s/%s/_changes?%s", db.Host, db.Database, args.Encode())
	r, err := http.Get(url)
	if err != nil {
		return nil, regularToCouchError(err)
	}
	if r.StatusCode != 200 {
		r.Body.Close()
		return nil, responseToCouchError(r)
	}
	j := json.NewDecoder(r.Body)
	go func() {
		defer close(c)
		defer r.Body.Close()
		for {
			var r DocRev
			err := j.Decode(&r)
			if err != nil {
				fmt.Printf("Error in json decoding: %s\n", err)
				return // nil, regularToCouchError(err)
			}
			if r.Seq == 0 {
				fmt.Printf("r.Seq == 0\n")
				return // nil, regularToCouchError(os.NewError(fmt.Sprintf("Sequence number was not set, or set to 0", r.Seq)))
			}
			c <- &r
		}
	}()
	return c, nil //regularToCouchError(os.NewError("This should be impossible to reach, just putting it here to shut up go"))
}

func (db *CouchDB) Info() (info *CouchInfo, cerr *CouchError) {
	info = new(CouchInfo)
	cerr = db.GetDocument(&info, "")
	if cerr != nil {
		return
	}
	return
}

func (db *CouchDB) Compact() (cerr *CouchError) {
	var s CouchSuccess
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/_compact", db.Host, db.Database), nil)
	if err != nil {
		return regularToCouchError(err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, cerr = couchDo(req, &s)
	if cerr != nil {
		return cerr
	}
	return nil
}

func (db *CouchDB) CompactView(designdoc string) (cerr *CouchError) {
	var s CouchSuccess
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/_compact/%s", db.Host, db.Database, designdoc), nil)
	if err != nil {
		return regularToCouchError(err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, cerr = couchDo(req, &s)
	if cerr != nil {
		return cerr
	}
	return nil
}
