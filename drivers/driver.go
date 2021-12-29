package drivers

import "database/sql"

type JoinType string

const (
	InnerJoin JoinType = "INNER JOIN"
	LeftJoin           = "LEFT JOIN"
	RightJoin          = "RIGHT JOIN"
)

type Driver interface {
	Column(tableAlias string, field string, modelAlias string, fieldAlias string) string
	From(schema string, table string, alias string) string
	Join(jType JoinType, schema string, table string, childAlias string, childField string, parentAlias string, parentField string) string
	Connect(Config) error
}

type Config struct {
	Host string
}

type driver struct {
	db sql.DB
}
