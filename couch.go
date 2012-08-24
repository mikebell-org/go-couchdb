package couchdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func Database(host, database, username, password string) (db *CouchDB, err error) {
	db = new(CouchDB)
	db.Host = host
	db.Database = database
	db.Username = username
	db.Password = password
	return db, nil
}

func CreateDatabase(host, database, username, password string) (*CouchDB, error) {
	var s CouchSuccess
	db, cerr := Database(host, database, username, password)
	if cerr != nil {
		return nil, cerr
	}
	req, err := db.request("PUT", "", nil)
	if err != nil {
		return nil, err
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

func (db *CouchDB) Delete() error {
	req, err := db.request("DELETE", "", nil)
	if err != nil {
		return err
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

func (db *CouchDB) GetDocument(doc interface{}, path string) error {
	return db.get(doc, path)
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

func (db *CouchDB) viewReq(design, view string, args ViewArgs, body io.Reader) (r *http.Request, err error) {
	var argstring, path string
	if argstring, err = args.Encode(); err != nil {
		return nil, err
	}
	if design == "" && view == "_all_docs" {
		path = fmt.Sprintf("_all_docs?%s", argstring)
	} else {
		path = fmt.Sprintf("_design/%s/_view/%s?%s", design, view, argstring)
	}
	if body == nil {
		return db.request("GET", path, body)
	} else {
		return db.request("POST", path, body)
	}
	panic("Should never be reached")
}

func (db *CouchDB) View(design, view string, args ViewArgs) (results *ViewResults, err error) {
	results = new(ViewResults)
	req, err := db.viewReq(design, view, args, nil)
	if err != nil {
		return nil, err
	}
	if _, err := couchDo(req, results); err != nil {
		return nil, err
	}
	return results, nil
}

func (db *CouchDB) PostView(design, view string, args ViewArgs, keys []interface{}) (results *ViewResults, err error) {
	results = new(ViewResults)
	r, errCh := jsonifyDoc(postViewData{Keys: keys})
	req, err := db.viewReq(design, view, args, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if _, err := couchDo(req, results); err != nil {
		return nil, err
	}
	if err := <-errCh; err != nil {
		return nil, err
	}
	return results, nil
}


func (db *CouchDB) ContinuousChanges(args url.Values) (chan *DocRev, error) {
	c := make(chan *DocRev)
	args.Set("feed", "continuous")
	url := fmt.Sprintf("_changes?%s", args.Encode())
	req, err := db.request("GET", url, nil)
	if err != nil {
		return nil, err
	}
	r, err := client.Do(req)
	if err != nil {
		return nil, err
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
			if err := j.Decode(&r); err != nil {
				return // nil, err
			}
			if r.Seq == 0 {
				return // nil, os.NewError(fmt.Sprintf("Sequence number was not set, or set to 0", r.Seq))
			}
			c <- &r
		}
	}()
	return c, nil //os.NewError("This should be impossible to reach, just putting it here to shut up go")
}

func (db *CouchDB) Info() (info *CouchInfo, cerr error) {
	info = new(CouchInfo)
	cerr = db.GetDocument(&info, "")
	if cerr != nil {
		return
	}
	return
}

func (db *CouchDB) Compact() (cerr error) {
	var s CouchSuccess
	req, err := db.request("POST", "_compact", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, cerr = couchDo(req, &s)
	if cerr != nil {
		return cerr
	}
	return nil
}

func (db *CouchDB) CompactView(designdoc string) (cerr error) {
	var s CouchSuccess
	req, err := db.request("POST", fmt.Sprintf("_compact/%s", designdoc), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, cerr = couchDo(req, &s)
	if cerr != nil {
		return cerr
	}
	return nil
}
