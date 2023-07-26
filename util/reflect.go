package util

import (
	"reflect"
)



func NewPtr(v interface{}) interface{} {
	type_ := reflect.TypeOf(v)
	switch type_.Kind() {
	case reflect.Ptr:
		return newPtr(type_.Elem())
	case reflect.Struct:
		return newPtr(type_)
	}
	return nil
}
func newPtr(type_ reflect.Type) interface{}{
	value:=reflect.New(type_)
	u:=value.Interface()
	return &u
}
