package drivers

import (
	"database/sql"
	"fmt"
	"github.com/Supersonido/sqlizer/queries"
)

type JoinType string

type Driver interface {
	Connect(Config) error
	Select(query queries.SelectQuery) (*sql.Rows, error)
	Insert(query queries.SelectQuery) (*sql.Row, error)
	Close()

	renderSelectColumn(queries.Column) string
	renderForm(queries.TableSource) string
	renderJoin(queries.Join, func() string) (string, []interface{})
	renderWhere([]queries.Where, string, func() string) (string, []interface{})
	renderColumnKey(queries.ColumnKey) string
	renderOrder([]queries.Order) string
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

type driverOptions struct {
	db *sql.DB
}

func (d driverOptions) Close() {
	err := d.db.Close()
	if err != nil {
		panic(err)
	}
}

func ValueSequence() func() string {
	num := 0
	return func() string {
		num += 1
		return fmt.Sprintf("$%d", num)
	}
}

func renderColumns(driver Driver, columns []queries.Column, prefix string) []string {
	var strColumns []string

	for _, column := range columns {
		var columnPrefix string
		if prefix == "" {
			columnPrefix = column.Alias
		} else {
			columnPrefix = fmt.Sprintf("%s.%s", prefix, column.Alias)
		}

		if column.Type == nil {
			strColumns = append(strColumns, renderColumns(driver, column.Nested, columnPrefix)...)
		} else {
			column.Alias = columnPrefix
			strColumns = append(strColumns, driver.renderSelectColumn(column))
		}
	}

	return strColumns
}

func renderJoins(driver Driver, joins []queries.Join, seq func() string) ([]string, []interface{}) {
	var strJoins []string
	var values []interface{}

	for _, join := range joins {
		j, v := driver.renderJoin(join, seq)
		strJoins = append(strJoins, j)
		values = append(values, v...)
	}

	return strJoins, values
}
