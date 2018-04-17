package couchdb

import (
	"encoding/json"
	"fmt"
)

type CouchInfo struct {
	Name               string      `json:"db_name"`
	DocCount           int         `json:"doc_count"`
	DocDelCount        int         `json:"doc_del_count"`
	UpdateSeq          interface{} `json:"update_seq"`
	PurgeSeq           int         `json:"purge_seq"`
	CompactRunning     bool        `json:"compact_running"`
	DiskSize           int         `json:"disk_size"`
	DataSize           int         `json:"data_size"`
	InstanceStartTime  string      `json:"instance_start_time"`
	DiskFormatVersion  int         `json:"disk_format_version"`
	CommittedUpdateSeq int         `json:"committed_update_seq"`
}

// Returns information about the database, can also be used to verify its existence
func (db *CouchDB) Info() (_ *CouchInfo, err error) {
	var info CouchInfo
	b, err := db.GetRaw("", "")
	if err != nil {
		return
	}
	var r = json.NewDecoder(b)
	if err = r.Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Starts compaction on a database
func (db *CouchDB) Compact() (err error) {
	var s CouchSuccess
	req, err := db.createRequest("POST", "_compact", "", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = couchDo(req, &s)
	if err != nil {
		return err
	}
	return nil
}

// Starts compaction on a view
func (db *CouchDB) CompactView(designdoc string) (err error) {
	var s CouchSuccess
	req, err := db.createRequest("POST", fmt.Sprintf("_compact/%s", designdoc), "", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = couchDo(req, &s)
	if err != nil {
		return err
	}
	return nil
}
