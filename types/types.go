package types

import (
	"reflect"
	"time"
)

type FieldType reflect.Type

// @TODO Update serializer when more types are added
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
	Float32Type            = FieldType(reflect.TypeOf((*float32)(nil)).Elem())
	Float32PtrType         = FieldType(reflect.TypeOf((*float32)(nil)))
	Float64Type            = FieldType(reflect.TypeOf((*float64)(nil)).Elem())
	Float64PtrType         = FieldType(reflect.TypeOf((*float64)(nil)))
)
