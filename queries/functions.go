package queries

import (
	"github.com/Supersonido/sqlizer/types"
	"reflect"
)

func Count(arg ...interface{}) *Function {
	return &Function{Operator: "count", Values: arg}
}

func CountDistinct(arg ...interface{}) *Function {
	return &Function{Operator: "countDist", Values: arg}
}

func Max(arg ...interface{}) *Function {
	return &Function{Operator: "max", Values: arg}
}

func Min(arg ...interface{}) *Function {
	return &Function{Operator: "min", Values: arg}
}

func RetypeFunction(fn *Function, t types.FieldType) *Function {
	rf := reflect.Type(t)
	fn.Type = &rf
	return fn
}
