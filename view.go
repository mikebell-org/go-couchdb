package couchdb

import ()

type UnescapedString string

var FalsePointer *bool
var TruePointer *bool

func init() {
	myFalse := false
	myTrue := true
	FalsePointer = &myFalse
	TruePointer = &myTrue
}

type postViewData struct {
	Keys []interface{} `json:"keys"`
}

type ViewArgs struct {
	Key            interface{}     `urlencode:"key"`
	Keys           []interface{}   `urlencode:"keys"`
	StartKey       interface{}     `urlencode:"startkey"`
	StartKey_DocID string          `urlencode:"startkey_docid"`
	EndKey         interface{}     `urlencode:"endkey"`
	EndKey_DocID   string          `urlencode:"endkey_docid"`
	Limit          uint            `urlencode:"limit"`
	Stale          UnescapedString `urlencode:"stale"` // Special string because we don't want to quote this one
	Descending     bool            `urlencode:"descending"`
	Skip           uint            `urlencode:"skip"`
	Group          bool            `urlencode:"group"`
	GroupLevel     uint            `urlencode:"group_level"`
	Reduce         *bool           `urlencode:"reduce"` // Because the default is true
	IncludeDocs    bool            `urlencode:"include_docs"`
	InclusiveEnd   *bool           `urlencode:"inclusive_end"` // Because the default is true
	UpdateSeq      bool            `urlencode:"update_seq"`
}

func (a *ViewArgs) Encode() (string, error) {
	return URLEncodeObject(*a)
}
