package types

import (
	"reflect"
	"time"
)

type FieldType reflect.Type

var (
	StringType             = FieldType(reflect.TypeOf((*string)(nil)).Elem())
	StringPtrType          = FieldType(reflect.TypeOf((*string)(nil)))
	IntegerType            = FieldType(reflect.TypeOf((*int)(nil)).Elem())
	IntegerPtrType         = FieldType(reflect.TypeOf((*int)(nil)))
	UnsignedIntegerType    = FieldType(reflect.TypeOf((*uint)(nil)).Elem())
	UnsignedIntegerPtrType = FieldType(reflect.TypeOf((*uint)(nil)))
	TimeType               = FieldType(reflect.TypeOf((*time.Time)(nil)).Elem())
	TimePtrType            = FieldType(reflect.TypeOf((*time.Time)(nil)))
	BooleanType            = FieldType(reflect.TypeOf((*bool)(nil)).Elem())
	BooleanPtrType         = FieldType(reflect.TypeOf((*bool)(nil)))
)
