package couchdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
)

var client http.Client

func init() {
	client = http.Client{Transport: &http.Transport{Dial: keepAliveDial}}
}

func keepAliveDial(nett, addr string) (c net.Conn, err error) {
	if c, err = net.Dial(nett, addr); err != nil {
		return
	}
	t, ok := c.(*net.TCPConn)
	if !ok {
		return c, fmt.Errorf("Socket returned by net.Dial is not a TCP socket, it's actually <%s>", c)
	}
	err = t.SetKeepAlive(true)
	return
}

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

func couchDo(req *http.Request, response interface{}) (int, error) {
	r, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer r.Body.Close()
	if r.StatusCode >= 300 {
		return r.StatusCode, responseToCouchError(r)
	}
	if response != nil {
		j := json.NewDecoder(r.Body)
		err = j.Decode(response)
		if err != nil {
			return r.StatusCode, err
		}
	}
	return r.StatusCode, nil
}
