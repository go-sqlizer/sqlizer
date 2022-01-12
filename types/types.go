package types

import (
	"reflect"
)

type FieldType reflect.Type

var (
	StingType           = FieldType(reflect.TypeOf((*string)(nil)).Elem())
	IntegerType         = FieldType(reflect.TypeOf((*int)(nil)).Elem())
	UnsignedIntegerType = FieldType(reflect.TypeOf((*uint)(nil)).Elem())
)
