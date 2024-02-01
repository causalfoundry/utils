package util

import (
	"fmt"
	"reflect"
	"strings"
)

func ExtractDBTags(input any) (tags []string, vals []any) {
	return ExtractTags(input, "db")
}

func ExtractTags(input any, tagName string) (tags []string, vals []any) {
	val := reflect.ValueOf(input)
	typ := reflect.TypeOf(input)

	if typ.Kind() != reflect.Struct {
		panic(fmt.Errorf("input not a struct: %T", input))
	}

	for i := 0; i < val.NumField(); i++ {
		va := val.Field(i)
		ty := typ.Field(i)

		isStruct := ty.Type.Kind() == reflect.Struct
		isSpecialType := strings.Contains("Time", ty.Type.Name())
		hasNoTag := ty.Tag == ""

		if isStruct && !isSpecialType && hasNoTag {
			t, v := ExtractTags(va.Interface(), tagName)
			tags = append(tags, t...)
			vals = append(vals, v...)
		} else {
			tags = append(tags, ty.Tag.Get(tagName))
			vals = append(vals, va.Interface())
		}
	}

	return
}

func StructDataType(s interface{}) []string {
	var ret []string
	T := reflect.TypeOf(s)
	if T.Kind() != reflect.Struct {
		return ret
	}
	for _, f := range reflect.VisibleFields(T) {
		if f.Type.Kind().String() == "slice" {
			ret = append(ret, "_"+strings.ToLower(f.Type.Elem().Name()))
		} else {
			ret = append(ret, strings.ToLower(f.Type.Name()))
		}
	}
	return ret
}

func StructTags(s interface{}, tagName string, except map[string]any) (ret []string) {
	T := reflect.TypeOf(s)
	if T.Kind() != reflect.Struct {
		return
	}
	for _, f := range reflect.VisibleFields(T) {
		name := f.Tag.Get(tagName)
		if name != "" {
			if v, ok := except[name]; ok {
				ret = append(ret, StructTags(v, tagName, nil)...)
			} else {
				ret = append(ret, f.Tag.Get(tagName))
			}
		}
	}
	return ret
}
