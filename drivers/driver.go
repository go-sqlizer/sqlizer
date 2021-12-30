package drivers

import (
	"database/sql"
	"fmt"
)

type JoinType string

const (
	InnerJoin JoinType = "INNER JOIN"
	LeftJoin           = "LEFT JOIN"
	RightJoin          = "RIGHT JOIN"
)

type Driver interface {
	SelectColumn(tableAlias string, field string, modelAlias string, fieldAlias string) string
	Column(tableAlias string, field string) string
	From(schema string, table string, alias string) string
	Join(jType JoinType, schema string, table string, childAlias string, childField string, parentAlias string, parentField string) string
	Connect(Config) error
	Close()
	Operator(action string) string
	Exec(query Query) error
	Query(query Query) (*sql.Rows, error)
}

type Config struct {
	Dialect         string
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSl             string
	ConnectionPool  int
	StartPoolOnBoot bool
}

type driver struct {
	db *sql.DB
}

type Query struct {
	Statement string
	Values    []interface{}
}

func ValuesSequence() func() string {
	num := 0
	return func() string {
		num += 1
		return fmt.Sprintf("$%d", num)
	}
}
