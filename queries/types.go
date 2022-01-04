package queries

import "reflect"

type Options struct {
	Logging func(...interface{})
	Where   []Where
	Limit   *int
	Offset  *int
	Include []Include
	Order   []Order
}

type Include struct {
	As       string
	Include  []Include
	Where    []Where
	JoinType JoinType
}

type PaginateOptions struct {
	Options
	Page    int
	PerPage int
}

type SelectQuery struct {
	Columns []Column
	From    TableSource
	Joins   []Join
	Values  []interface{}
	Options
}

type Column struct {
	Alias        string
	Nested       []Column
	Type         *reflect.Type
	Source       ColumnSource
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
	Key      ColumnKey
	Value    interface{}
	Nested   []Where
	Operator string
}

type OrderType uint8

const (
	DescOrder OrderType = iota
	AscOrder
)

type Order struct {
	Key  ColumnKey
	Type OrderType
}
