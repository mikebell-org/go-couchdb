package couchdb

import (
	"encoding/json"
	"io"
	"net/http"
)

var client http.Client

func jsonifyDoc(doc interface{}) (io.Reader, chan error) {
	errCh := make(chan error)
	r, w := io.Pipe()
	go func() {
		j := json.NewEncoder(w)
		err := j.Encode(doc)
		w.Close()
		errCh <- err
		close(errCh)
	}()
	return r, errCh
}

func couchDo(req *http.Request, response interface{}) (int, *CouchError) {
	r, err := client.Do(req)
	if err != nil {
		return 0, regularToCouchError(err)
	}
	if r.StatusCode >= 300 {
		return r.StatusCode, responseToCouchError(r)
	}
	if response != nil {
		j := json.NewDecoder(r.Body)
		err = j.Decode(response)
		if err != nil {
			return r.StatusCode, regularToCouchError(err)
		}
	}
	return r.StatusCode, nil
}
