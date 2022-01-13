package queries

import (
	"github.com/Supersonido/sqlizer/types"
	"reflect"
)

func Count(arg interface{}) *Function {
	t := reflect.Type(types.UnsignedIntegerType)
	return &Function{Operator: "count", Values: []interface{}{arg}, Type: &t}
}

func CountDistinct(arg interface{}) *Function {
	t := reflect.Type(types.UnsignedIntegerType)
	return &Function{Operator: "countDist", Values: []interface{}{arg}, Type: &t}
}

func Max(t reflect.Type, arg ...interface{}) *Function {
	return &Function{Operator: "max", Values: arg, Type: &t}
}

func Min(t reflect.Type, arg ...interface{}) *Function {
	return &Function{Operator: "min", Values: []interface{}{arg}, Type: &t}
}
