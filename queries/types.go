package queries

import (
	"reflect"
)

type QueryOptions struct {
	Logging func(...interface{})
	Where   []Where
	Limit   *int
	Offset  *int
	Include []Include
	Order   []Order
	Fields  Fields
	Group   []ColumnKey
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
	Page    int
	PerPage int
}

type SelectQuery struct {
	Columns []Column
	From    TableSource
	Joins   []Join
	Values  []interface{}
	QueryOptions
}

type Column struct {
	Alias        string
	Nested       *[]Column
	Type         *reflect.Type
	Source       *ColumnSource
	Function     *Function
	IsPrimaryKey bool
}

type ColumnSource struct {
	Alias string
	Field string
}

type ColumnKey struct {
	Alias string
	Field string
}

func (ck ColumnKey) ToSQL(serializer SQLSerializer) string {
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
