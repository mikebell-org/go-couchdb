package couchdb

type ViewDef struct{
	Map string
	Reduce string
}

type Json map[string]interface{}

type ViewDefMap map[string]*ViewDef

/**
	Handles defining and syncing design documents

	For now only supports map-reduce views
*/
type DesignDocument struct{
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

func (dd *DesignDocument) toJson() (out Json) {
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

