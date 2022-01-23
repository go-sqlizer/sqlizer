package types

import (
	"reflect"
)

type FieldType reflect.Type

var (
	StingType           = FieldType(reflect.TypeOf((*string)(nil)).Elem())
	StingPtrType        = FieldType(reflect.TypeOf((*string)(nil)))
	IntegerType         = FieldType(reflect.TypeOf((*int)(nil)).Elem())
	IntegerPtrType      = FieldType(reflect.TypeOf((*int)(nil)))
	UnsignedIntegerType = FieldType(reflect.TypeOf((*uint)(nil)).Elem())
)
