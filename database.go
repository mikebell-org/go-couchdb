package couchdb

// Opens a database
func Database(host, database, username, password string) (db *CouchDB, err error) {
	db = new(CouchDB)
	db.Host = host
	db.Database = database
	db.Username = username
	db.Password = password
	return db, nil
}

// Creates a new database and returns the struct as Database() would.
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

// Deletes the database in question. Scary!
func (db *CouchDB) DeleteDatabase() error {
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
