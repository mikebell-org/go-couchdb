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

func Database(host, database, username, password string) (db *CouchDB, err *CouchError) {
	db = new(CouchDB)
	db.Host = host
	db.Database = database
	db.Username = username
	db.Password = password
	return db, nil
}

func CreateDatabase(host, database, username, password string) (*CouchDB, *CouchError) {
	var s CouchSuccess
	db, cerr := Database(host, database, username, password)
	if cerr != nil {
		return nil, cerr
	}
	req, err := db.request("PUT", "", nil)
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

func (db *CouchDB) get(doc interface{}, path string) *CouchError {
	req, err := db.request("GET", path, nil)
	if err != nil {
		fmt.Printf("go-couchdb: Failed creating request: %s\n", err)
		return regularToCouchError(err)
	}
	code, cerr := couchDo(req, doc)
	if cerr != nil {
		fmt.Printf("go-couchdb: Failed in couchDo: %s\n", cerr)
		return cerr
	}
	if code != 200 {
		// FIXME Unexpected code. Do something?
	}
	fmt.Printf("go-couchdb: Successfully got document\n")
	return nil
}

func (db *CouchDB) Delete() *CouchError {
	req, err := db.request("DELETE", "", nil)
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
	req, err := db.request("GET", path, nil)
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
	req, err := db.request("PUT", path, r)
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
	req, err := db.request("POST", "", r)
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

func (db *CouchDB) BulkUpdate(c *BulkCommit) (*BulkCommitResponse, *CouchError) {
	var s BulkCommitResponse
	r, errCh := jsonifyDoc(c)
	req, err := db.request("POST", "_bulk_docs", r)
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
	req, err := db.request("DELETE", fmt.Sprintf("%s?rev=%s", path, rev), nil)
	if err != nil {
		fmt.Printf("Returning error from request creation: %s\n", err)
		return nil, regularToCouchError(err)
	}
	code, cerr := couchDo(req, &s)
	if cerr != nil {
		fmt.Printf("Error in couchDo: %s\n", cerr)
		return nil, cerr
	}
	if code != 200 {
		// FIXME Unexpected code. Do something?
	}
	fmt.Printf("Deleted successfully\n")
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
	url := fmt.Sprintf("_changes?%s", args.Encode())
	req, err := db.request("GET", url, nil)
	if err != nil {
		return nil, regularToCouchError(err)
	}
	r, err := client.Do(req)
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
	req, err := db.request("POST", "_compact", nil)
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
	req, err := db.request("POST", fmt.Sprintf("_compact/%s", designdoc), nil)
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
