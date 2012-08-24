package couchdb

import (
	"encoding/json"
	"log"
	"reflect"
)

type ViewDef struct {
	Map    string
	Reduce string
}

type Json map[string]interface{}

type ViewDefMap map[string]*ViewDef

/**
Handles defining and syncing design documents

For now only supports map-reduce views
*/
type DesignDocument struct {
	// The name of the design document, where the id is set to _design/<name>
	Name string

	// named map-rdeuce functions
	Views ViewDefMap
}

func NewDesignDocument(name string) *DesignDocument {
	obj := new(DesignDocument)
	obj.Name = name
	obj.Views = make(ViewDefMap)
	return obj
}

func (dd *DesignDocument) SetMap(name, map_fn string) {
	view, has_it := dd.Views[name]
	if !has_it {
		view = new(ViewDef)
	}
	view.Map = map_fn
	dd.Views[name] = view
}

func (dd *DesignDocument) SetMapReduce(name, map_fn, reduce_fn string) {
	dd.SetMap(name, map_fn)
	dd.Views[name].Reduce = reduce_fn
}

func (dd *DesignDocument) ToJson() (out Json) {
	out = make(Json)
	out["_id"] = "_design/" + dd.Name
	out["language"] = "javascript"
	json_views := make(Json)
	for name, view := range dd.Views {
		json_view := make(Json)
		if view.Map != "" {
			json_view["map"] = view.Map
		}
		if view.Reduce != "" {
			json_view["reduce"] = view.Reduce
		}
		json_views[name] = json_view
	}
	out["views"] = json_views
	return
}

// forcibly sync the design document to the database,
// overriding any existing design document with the same name
func (dd *DesignDocument) ForceSync(db *CouchDB) error {
	doc := dd.ToJson()
	path := "/" + doc["_id"].(string)
	orig := make(Json)

	// get the document to get the rev
	// this is the part where forcible override existing document!
	db.GetDocument(&orig, path)
	if orig["_rev"] != nil {
		doc["_rev"] = orig["_rev"]
	}

	// check if there's any change before saving
	// the 'orig' won't have the _id so set it before the comparison ..
	orig["_id"] = doc["_id"]
	if JsonEqual(doc, orig) {
		// no change, so saving is pointless (it'll only increment the _rev property)
		return nil
	}

	log.Println("Force-Syncing design document:", path)

	_, err := db.PutDocument(doc, path)
	if err != nil {
		log.Println("Failed to sync design document:", dd.Name, "\n", err)
		return err
	}
	return nil
}

func JsonEqual(obj1, obj2 Json) bool {
	// reflect.DeepEqual fails because we use type aliases sometimes (e.g. Json instead of map[string]interface{} directly)
	// so as a hack, convert both objects to json strings, then convert the json strings back to objects
	// and compare these objects instead
	// it's a trick to normalize both objects so they get the same type

	normalize := func(obj Json) (out Json) {
		json_text, _ := json.Marshal(obj)
		out = make(Json)
		json.Unmarshal(json_text, &out)
		return
	}

	obj1 = normalize(obj1)
	obj2 = normalize(obj2)

	return reflect.DeepEqual(obj1, obj2)
}
