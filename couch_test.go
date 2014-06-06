package couchdb

import (
	"fmt"
	"testing"
	"time"
)

type testdoc struct {
	BasicDocumentWithMtime
	Test string
}

var WeirdDocIDs = []string{
	"abc",
	"Hello World",
	"Hello/World",
	"4$44$^&",
	"_design/this_is_just_a_test",
	"+1",
	"^%$@*!@&*((",
	"tab	",
}

func TestMain(t *testing.T) {
	var doc testdoc
	var change *DocRev
	var ok bool
	var results *AllDocsResult

	doc.Test = "Hello World!"
	db, err := CreateDatabase("http://127.0.0.1:5984", "go_couchdb_test_suite", "", "")
	if err != nil {
		t.Fatalf("Error creating new database for testing: %s.\nNote, tests expect a couch database on 127.0.0.1:5984, anyone have better ideas?", err)
	}
	fmt.Printf("Stage 1 complete\n")
	defer db.DeleteDatabase()

	args := ChangesArgs{Heartbeat: 30000, Since: 0, Feed: "continuous"}
	changeChan, errChan := db.ContinuousChanges(args)
	if changeChan == nil {
		fmt.Printf("DEBUG: Error initializing changes feed\n")
		t.Fatalf("Error initializing changes feed: %s", <-errChan)
	}

	changeOrErr := func() *DocRev {
		select {
		case change, ok = <-changeChan:
			if !ok {
				t.Fatal("Error on continuous changes feed")
			}
			return change
		case err := <-errChan:
			if err != nil {
				t.Fatalf("Error on continuous changes feed: %s", err)
			}
		case _ = <-time.After(5 * time.Second):
			t.Fatalf("Timeout waiting for a changes event that was anticipated")
		}
		panic("Should never be reached")
	}

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
	change = changeOrErr()
	/*
		if change, ok = <-c; !ok {
			t.Fatalf("Error from changes feed: %s", <-errChan)
		}
	*/
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
		t.Fatalf("Error updating existing doc (%+v): %s", doc, err)
	}
	if PutSuccess.Rev == PostSuccess.Rev {
		t.Fatalf("Error updating a doc, rev stayed the same?")
	}
	fmt.Printf("Stage 4 complete\n")

	change = changeOrErr()
	if change.ID != PostSuccess.ID {
		t.Errorf("Change I got from the changes API didn't match what I got from my POST")
	}

	a := ViewArgs{Reduce: FalsePointer, IncludeDocs: true, Limit: 4}

	if results, err = db.AllDocs(a); err != nil {
		t.Errorf("Failed calling _all_docs view: %s", err)
	}
	fmt.Printf("%+v\n", results)
	fmt.Printf("Stage 5 complete\n")

	if _, err = db.DeleteDocument(PostSuccess.ID, PutSuccess.Rev); err != nil {
		t.Fatalf("Error deleting doc: %s", err)
	}
	fmt.Printf("Stage 6 complete\n")

	change = changeOrErr()
	if change.ID != PostSuccess.ID {
		t.Errorf("Change I got from the changes API didn't match what I got from my POST")
	}

	// Last test, big bulk commit of weird docids followed by getting each one individually to ensure docids are being encoded correctly
	bc := BulkCommit{}
	for _, docid := range WeirdDocIDs {
		td := testdoc{}
		td.ID = docid
		bc.Docs = append(bc.Docs, td)
	}
	fmt.Printf("Bulk committing: %+v\n", bc)
	if bc_response, err := db.BulkUpdate(&bc); err != nil {
		t.Fatalf("Error doing a bulk addition of weird docids: %+v %s", bc_response, err)
	} else {
		var errorStrings []string
		for _, row := range *bc_response {
			if row.Error != "" {
				errorStrings = append(errorStrings, row.Error)
			}
		}
		if len(errorStrings) != 0 {
			t.Fatalf("Error writing one or more weird docids: %s", errorStrings)
		}
	}
	for _, docid := range WeirdDocIDs {
		var doc testdoc
		if err = db.GetDocument(&doc, docid); err != nil {
			t.Fatalf("Error getting one of the weird docid docs: %s", err)
		}
	}

	// All done, delete the DB to clean up and as a final test
	err = db.DeleteDatabase()
	if err != nil {
		t.Fatalf("Error deleting database: %s", err)
	}
}
