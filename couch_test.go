package couchdb

import (
	"fmt"
	"net/url"
	//	"runtime"
	//	"time"
	"testing"
)

type testdoc struct {
	ID   string `json:"_id,omitempty"`
	Rev  string `json:"_rev,omitempty"`
	Test string
}

func TestMain(t *testing.T) {
	a := ViewArgs{Key: "mike", Reduce: FalsePointer, IncludeDocs: true, Limit: 4}
	str, err := a.Encode()
	if err != nil {
		t.Fatalf("Error encoding view URL: %s", err)
	}
	fmt.Printf("Encoding of view URL would be: %s\n", str)

	var doc testdoc
	var change *DocRev
	doc.Test = "Hello World!"
	db, err := CreateDatabase("http://127.0.0.1:5984", "go_couchdb_test_suite", "", "")
	if err != nil {
		t.Fatalf("Error creating new database for testing: %s.\nNote, tests expect a couch database on 127.0.0.1:5984, anyone have better ideas?", err)
	}
	fmt.Printf("Stage 1 complete\n")
	defer db.Delete()

	//	c := make(chan *DocRev)
	args := make(url.Values)
	args.Set("keepalive", "30000")
	args.Set("since", "0")
	c, err := db.ContinuousChanges(args)

	PostSuccess, err := db.PostDocument(doc)
	if err != nil {
		t.Fatalf("Error creating new doc using POST: %s", err)
	}
	if PostSuccess == nil {
		t.Fatalf("Somehow success == nil but err == nil!")
	}
	if PostSuccess.ID == "" {
		t.Fatalf("Didn't get a DocID back from our POST")
	}
	fmt.Printf("Stage 2 complete\n")
	change = <-c
	if change.ID != PostSuccess.ID {
		t.Errorf("Change I got from the changes API didn't match what I got from my POST")
	}

	err = db.GetDocument(&doc, PostSuccess.ID)
	if err != nil {
		t.Fatalf("Error retrieving the doc we just made using POST: %s", err)
	}
	if doc.Test != "Hello World!" {
		t.Fatalf("Retreived doc doesn't match the one we just POSTed!")
	}
	fmt.Printf("Stage 3 complete\n")

	PutSuccess, err := db.PutDocument(&doc, PostSuccess.ID)
	if err != nil {
		t.Fatalf("Error updating existing doc: %s", err)
	}
	if PutSuccess.Rev == PostSuccess.Rev {
		t.Fatalf("Error updating a doc, rev stayed the same?")
	}
	fmt.Printf("Stage 4 complete\n")

	change = <-c
	if change.ID != PostSuccess.ID {
		t.Errorf("Change I got from the changes API didn't match what I got from my POST")
	}

	_, err = db.DeleteDocument(PostSuccess.ID, PutSuccess.Rev)
	if err != nil {
		t.Fatalf("Error deleting doc: %s", err)
	}
	fmt.Printf("Stage 5 complete\n")

	change = <-c
	if change.ID != PostSuccess.ID {
		t.Errorf("Change I got from the changes API didn't match what I got from my POST")
	}

	err = db.Delete()
	if err != nil {
		t.Fatalf("Error deleting database: %s", err)
	}
}
