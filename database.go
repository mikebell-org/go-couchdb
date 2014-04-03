package couchdb

import (
	"errors"
	"fmt"
	"strings"
)

var slashReplacer *strings.Replacer

func init() {
	slashReplacer = strings.NewReplacer("/", "%2F")
}

func validDBname(database string) error {
	if len(database) == 0 {
		return errors.New("Database name cannot be blank")
	}
	if database[0] < 'a' || database[0] > 'z' {
		return errors.New("First character of database name must be in the range of a-z")
	}
	slashReplacer.Replace(database)
	for _, c := range database {
		switch {
		case c == '/':
			continue
		case '0' <= c && c <= '9':
			continue
		case 'a' <= c && c <= 'z':
			continue
		case c == '_' || c == '$' || c == '(' || c == ')' || c == '+' || c == '-':
			continue
		default:
			return fmt.Errorf("Invalid character %s in database name", c)
		}
	}
	return nil
}

// Opens a database
func Database(host, database, username, password string) (db *CouchDB, err error) {
	db = new(CouchDB)
	db.Host = host
	db.Username = username
	db.Password = password
	if err = validDBname(database); err != nil {
		return nil, err
	}
	db.Database = slashReplacer.Replace(database)
	return db, nil
}

// Creates a new database and returns the struct as Database() would.
func CreateDatabase(host, database, username, password string) (*CouchDB, error) {
	var s CouchSuccess
	db, cerr := Database(host, database, username, password)
	if cerr != nil {
		return nil, cerr
	}
	req, err := db.createRequest("PUT", "", "", nil)
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

// Deletes the database in question. Scary!
func (db *CouchDB) DeleteDatabase() error {
	req, err := db.createRequest("DELETE", "", "", nil)
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
