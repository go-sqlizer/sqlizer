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
	return &Function{Operator: "countDist", Values: []interface{}{arg}}
}

func Max(arg interface{}) *Function {
	return &Function{Operator: "max", Values: []interface{}{arg}}
}

func Min(arg interface{}) *Function {
	return &Function{Operator: "min", Values: []interface{}{arg}}
}
