package queries

import (
	"fmt"
	"github.com/Supersonido/sqlizer"
	"strings"
)

type WhereOption string

func Eq(tableAlias string, field string, value interface{}) WhereOption {
	fullField := sqlizer.Conn.Column(tableAlias, field)
	return WhereOption(fmt.Sprintf(`%s = %s`, fullField, valueCast(value)))
}

func NotEq(tableAlias string, field string, value interface{}) WhereOption {
	fullField := sqlizer.Conn.Column(tableAlias, field)
	return WhereOption(fmt.Sprintf(`%s != %s`, fullField, valueCast(value)))
}

func Or(where ...WhereOption) WhereOption {
	var fullWhere []string
	for _, w := range where {
		fullWhere = append(fullWhere, string(w))
	}

	return WhereOption(fmt.Sprintf("(%s)", strings.Join(fullWhere, " OR ")))
}

func And(where ...WhereOption) WhereOption {
	var fullWhere []string
	for _, w := range where {
		fullWhere = append(fullWhere, string(w))
	}

	return WhereOption(fmt.Sprintf("(%s)", strings.Join(fullWhere, " AND ")))
}

func valueCast(v interface{}) string {
	switch v.(type) {
	case string:
		return fmt.Sprintf(`'%s'`, v)
	default:
		return fmt.Sprintf(`%s`, v)
	}
}
