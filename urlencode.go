package couchdb

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// This type is for strings that shouldn't be JSON-encoded in URL-encoded structs. For example feed=continuous instead of the default feed="continuous"
type UnescapedString string

// Parameters like Reduce and InclusiveEnd default to true and thus are set to *bool instead of bool, providing a tri-state (true, false, unset-which-defaults-to-True). Thus you'll need to use these instead of just true and false for those values.
var FalsePointer *bool
var TruePointer *bool

func init() {
	myFalse := false
	myTrue := true
	FalsePointer = &myFalse
	TruePointer = &myTrue
}

func urlEncodeObject(a interface{}) (string, error) {
	v := reflect.ValueOf(a)
	t := reflect.TypeOf(a)
	numElements := v.NumField()
	parts := make([]string, 0, numElements)
	for i := 0; i < numElements; i++ {
		vf := v.Field(i)
		if isEmptyValue(vf) {
			continue
		}
		value, err := fieldValue(vf)
		if err != nil {
			return "", err
		}
		key := t.Field(i).Tag.Get("urlencode")
		parts = append(parts, key+"="+value)
	}
	return strings.Join(parts, "&"), nil
}

func fieldValue(v reflect.Value) (string, error) {
	if !v.CanInterface() {
		return "", fmt.Errorf("Error in viewargs: cannot show %s as an interface", v)
	}
	i := v.Interface()
	if str, ok := i.(UnescapedString); ok {
		return string(str), nil
	}
	b, err := json.Marshal(i)
	if err != nil {
		return "", fmt.Errorf("Error %s encoding value %s of viewargs\n", err, v)
	}
	return url.QueryEscape(string(b)), nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
