package drivers

import (
	"database/sql"
	"fmt"
	"github.com/Supersonido/sqlizer/queries"
	"strings"
	"time"
)

type Postgres struct {
	driverOptions
}

func (p *Postgres) Connect(config Config) error {
	connString := p.connectionString(config)
	db, err := sql.Open(config.Dialect, connString)
	if err != nil {
		return err
	}

	if config.ConnectionPool > 0 {
		db.SetMaxOpenConns(config.ConnectionPool)
		db.SetMaxIdleConns(config.ConnectionPool)
	}

	if config.StartPoolOnBoot {
		for i := 0; i < config.ConnectionPool; i++ {
			err = db.Ping()
			if err != nil {
				return err
			}
		}
	}

	p.db = db
	return nil
}

func (p *Postgres) Select(query queries.SelectQuery) (*sql.Rows, error) {
	var extra []string
	valueSequence := ValueSequence()
	columns := renderColumns(p, query.Columns, "")
	joins, values := renderJoins(p, query.Joins, valueSequence)

	if len(query.Where) > 0 {
		where, newValues := p.renderWhere(query.Where, "AND", valueSequence)
		extra = append(extra, fmt.Sprintf("WHERE %s", where))
		values = append(values, newValues...)
	}

	statement := fmt.Sprintf(
		"SELECT %s \nFROM %s\n%s\n%s",
		strings.Join(columns, ",\n\t"),
		p.renderForm(query.From),
		strings.Join(joins, "\n"),
		strings.Join(extra, "\n"),
	)

	if query.Logging != nil {
		fmt.Println(statement, values)
	}

	start := time.Now()
	rows, r := p.db.Query(statement, values...)
	fmt.Printf("\n\nQuery exec took %s\n", time.Since(start))
	return rows, r
}

func (p *Postgres) Insert(query queries.SelectQuery) (*sql.Row, error) {
	return nil, nil
}

func (p *Postgres) connectionString(config Config) string {
	return fmt.Sprintf(
		`host=%s port=%d user=%s password=%s dbname=%s sslmode=%s`,
		config.Host, config.Port, config.User, config.Password, config.Name, config.SSl)

}

func (p *Postgres) renderSelectColumn(column queries.Column) string {
	if column.Source.Alias == "" {
		return fmt.Sprintf(`"%s" AS "%s"`, column.Source.Field, column.Alias)
	}

	return fmt.Sprintf(`"%s"."%s" AS "%s"`, column.Source.Alias, column.Source.Field, column.Alias)
}

func (p *Postgres) renderForm(table queries.TableSource) string {
	schema := table.Schema
	if schema == "" {
		schema = "public"
	}

	return fmt.Sprintf(`"%s"."%s" AS "%s"`, schema, table.Table, table.Alias)
}

func (p *Postgres) renderJoin(join queries.Join, seq func() string) (string, []interface{}) {
	var joinTypeStr string
	switch join.Type {
	case queries.InnerJoin:
		joinTypeStr = "INNER JOIN"
	case queries.LeftJoin:
		joinTypeStr = "LEFT JOIN"
	case queries.RightJoin:
		joinTypeStr = "RIGHT JOIN"
	}

	where, values := p.renderWhere(join.Where, "AND", seq)
	return fmt.Sprintf(`%s %s ON %s`, joinTypeStr, p.renderForm(join.To), where), values
}

func (p *Postgres) renderColumnKey(key queries.ColumnKey) string {
	return fmt.Sprintf(`"%s"."%s"`, key.Alias, key.Field)
}

func (p *Postgres) renderWhere(wheres []queries.Where, linker string, seq func() string) (string, []interface{}) {
	var values []interface{}
	var filters []string

	for _, where := range wheres {
		key := p.renderColumnKey(where.Key)

		value := func() string {
			switch where.Value.(type) {
			case queries.ColumnKey:
				return p.renderColumnKey(where.Value.(queries.ColumnKey))
			default:
				values = append(values, where.Value)
				return seq()
			}
		}

		switch where.Operator {
		case "=", "!=":
			filters = append(filters, fmt.Sprintf("%s %s %s", key, where.Operator, value()))
		case "and":
			newFilters, NewValues := p.renderWhere(where.Nested, "AND", seq)
			filters = append(filters, fmt.Sprintf("(%s)", newFilters))
			values = append(values, NewValues...)
		case "or":
			newFilters, NewValues := p.renderWhere(where.Nested, "OR", seq)
			filters = append(filters, fmt.Sprintf("(%s)", newFilters))
			values = append(values, NewValues...)
		}
	}

	if len(filters) == 0{
		filters = append(filters, "TRUE")
	}

	return strings.Join(filters, fmt.Sprintf(" %s ", linker)), values
}
