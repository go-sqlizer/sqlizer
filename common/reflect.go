package common

import "reflect"

func TypeResolver(p reflect.Type) *reflect.Type {
	switch p.Kind() {
	case reflect.Ptr, reflect.Array, reflect.Slice:
		return TypeResolver(p.Elem())
	default:
		return &p
	}
}

func ValueFinder(result *reflect.Value) *reflect.Value {
	switch result.Kind() {
	case reflect.Ptr:
		newResult := result.Elem()
		return &newResult
	}

	return result
}
