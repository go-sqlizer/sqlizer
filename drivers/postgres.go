package drivers

import (
	"database/sql"
	"fmt"
	"github.com/go-sqlizer/sqlizer/queries"
	"strings"
)

type Postgres struct {
	CommonDriver
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

	p.CommonDriver = CommonDriver{
		db:                db,
		serializer:        queries.SQLSerializer(p),
		WhereOperators:    whereOperators,
		JoinOperators:     joinOperators,
		OrderOperators:    orderOperators,
		FunctionOperators: functionOperators,
	}

	return nil
}

func (p *Postgres) Insert(insert queries.BasicQuery) (sql.Result, error) {
	var columns []string
	var valueSeq []string
	var values []interface{}
	seq := valueSequence()

	for _, column := range insert.Columns {
		columns = append(columns, column.Source.Field)
		valueSeq = append(valueSeq, seq())
		values = append(values, column.Value)
	}

	statement := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s);",
		p.SerializeTableSource(insert.From),
		strings.Join(columns, ", "),
		strings.Join(valueSeq, ", "),
	)

	if insert.Logging != nil {
		fmt.Println(statement, values)
	}

	if insert.Transaction != nil {
		return insert.Transaction.Exec(statement, values...)
	} else {
		return p.db.Exec(statement, values...)
	}
}

func (p *Postgres) InsertReturning(insert queries.BasicQuery) *sql.Row {
	var columns []string
	var valueSeq []string
	var values []interface{}
	var returning []string
	seq := valueSequence()

	for _, column := range insert.Columns {
		columns = append(columns, column.Source.Field)
		valueSeq = append(valueSeq, seq())
		values = append(values, column.Value)
		returning = append(returning, p.SerializeColumn(column))
	}

	statement := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING %s;",
		p.SerializeTableSource(insert.From),
		strings.Join(columns, ", "),
		strings.Join(valueSeq, ", "),
		strings.Join(returning, ", "),
	)

	if insert.Logging != nil {
		fmt.Println(statement, values)
	}

	if insert.Transaction != nil {
		return insert.Transaction.QueryRow(statement, values...)
	} else {
		return p.db.QueryRow(statement, values...)
	}
}

func (_ Postgres) connectionString(config Config) string {
	return fmt.Sprintf(
		`host=%s port=%d user=%s password=%s dbname=%s sslmode=%s`,
		config.Host, config.Port, config.User, config.Password, config.Name, config.SSl)

}

func (_ Postgres) SerializeColumn(column queries.Column) string {
	if column.Source.Alias == "" {
		return fmt.Sprintf(`"%s" AS "%s"`, column.Source.Field, column.Alias)
	}

	return fmt.Sprintf(`"%s"."%s" AS "%s"`, column.Source.Alias, column.Source.Field, column.Alias)
}

func (_ Postgres) SerializeTableSource(table queries.TableSource) string {
	schema := table.Schema
	if schema == "" {
		schema = "public"
	}

	if table.Alias == "" {
		return fmt.Sprintf(`"%s"."%s"`, schema, table.Table)
	}

	return fmt.Sprintf(`"%s"."%s" AS "%s"`, schema, table.Table, table.Alias)
}

func (_ Postgres) SerializeColumnKey(key queries.ColumnKey) string {
	return fmt.Sprintf(`"%s"."%s"`, key.Alias, key.Field)
}

func (_ Postgres) SerializeAlias(raw string, alias string) string {
	return fmt.Sprintf(`%s AS "%s"`, raw, alias)
}
