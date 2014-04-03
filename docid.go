package couchdb

import (
	"net/url"
)

func escape_docid(docid string) string {
	return url.QueryEscape(docid)
}
