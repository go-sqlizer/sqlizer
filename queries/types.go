package queries

import (
	"github.com/go-sqlizer/sqlizer/types"
	"reflect"
)

type QueryOptions struct {
	Where       []Where
	Limit       *int
	Offset      *int
	Include     []Include
	Order       []Order
	Fields      Fields
	Group       []ColumnValue
	Logging     func(...interface{})
	Transaction types.Transaction
}

type InsertOptions struct {
	Logging     func(...interface{})
	Transaction types.Transaction
}

type UpdateOptions struct {
	Logging     func(...interface{})
	Transaction types.Transaction
	Where       []Where
}

type DeleteOptions struct {
	Logging     func(...interface{})
	Transaction types.Transaction
	Where       []Where
}

type Include struct {
	As       string
	Include  []Include
	Where    []Where
	JoinType JoinType
	Fields   Fields
}

type Fields struct {
	Includes []Field
	Excludes []string
}

type Field struct {
	As string
	Fn *Function
}

type PaginateOptions struct {
	QueryOptions
	Page    uint
	PerPage uint
}

type PaginateResults struct {
	Total   uint
	Pages   uint
	Page    uint
	PerPage uint
}

type BasicQuery struct {
	Columns []Column
	From    TableSource
	Joins   []Join
	QueryOptions
}

type InsertQuery struct {
	Columns   []Column
	From      TableSource
	Values    []interface{}
	Returning *reflect.Value
	InsertOptions
}

type Column struct {
	Alias        string
	Nested       *[]Column
	Type         *reflect.Type
	ColumnType   *reflect.Type
	Source       *ColumnSource
	Function     *Function
	IsPrimaryKey bool
	Get          func(value interface{}) interface{}
	Set          func(value interface{}) interface{}
	Value        interface{}
}

type ColumnSource struct {
	Alias string
	Field string
}

type ColumnValue struct {
	Alias string
	Field string
}

func (ck ColumnValue) ToSQL(serializer SQLSerializer) string {
	return serializer.SerializeColumnKey(ck)
}

type JoinType uint8

const (
	InnerJoin JoinType = iota
	LeftJoin
	RightJoin
)

type Join struct {
	Type  JoinType
	From  string
	To    TableSource
	Where []Where
}

type TableSource struct {
	Schema string
	Table  string
	Alias  string
}

type Where struct {
	Key      SQLRender
	Value    interface{}
	Operator string
}

type Function struct {
	Operator string
	Values   []interface{}
	Type     *reflect.Type
}

type OrderType uint8

const (
	DescOrder OrderType = iota
	AscOrder
)

type Order struct {
	Key  SQLRender
	Type OrderType
}
