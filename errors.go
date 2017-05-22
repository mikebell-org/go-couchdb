package couchdb

import "fmt"

var MissingDocumentIDError = fmt.Errorf("Empty document ID not valid here")
