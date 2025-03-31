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
	db, err := sql.Open(config.Dialect, config.Url)
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

func (p *Postgres) InsertReturning(insert queries.BasicQuery) *sql.Row {
	var columns []string
	var valueSeq []string
	var values []interface{}
	var returning []string
	seq := valueSequence()

	for _, column := range insert.Columns {
		columns = append(columns, p.SerializeColumn(column))
		valueSeq = append(valueSeq, seq())
		values = append(values, column.Value)
		returning = append(returning, p.SerializeColumnAlias(column))
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

func (p *Postgres) UpdateReturning(update queries.BasicQuery) *sql.Row {
	var columns []string
	var values []interface{}
	var extra []string
	var returning []string
	seq := valueSequence()

	for _, column := range update.Columns {
		returning = append(returning, p.SerializeColumnAlias(column))
		if column.Value != nil {
			columns = append(columns, fmt.Sprintf("%s = %s", p.SerializeColumn(column), seq()))
			values = append(values, column.Value)
		}
	}

	if len(update.Where) > 0 {
		where, newValues := p.WhereOperators["and"](nil, update.Where, &p.CommonDriver, seq)
		extra = append(extra, fmt.Sprintf("WHERE %s", where))
		values = append(values, newValues...)
	}

	statement := fmt.Sprintf(
		"UPDATE %s SET %s %s RETURNING %s;",
		p.serializer.SerializeTableSource(update.From),
		strings.Join(columns, ", "),
		strings.Join(extra, "\n"),
		strings.Join(returning, ", "),
	)

	if update.Logging != nil {
		fmt.Println(statement, values)
	}

	if update.Transaction != nil {
		return update.Transaction.QueryRow(statement, values...)
	} else {
		return p.db.QueryRow(statement, values...)
	}
}

func (_ *Postgres) SerializeColumnAlias(column queries.Column) string {
	if column.Source.Alias == "" {
		return fmt.Sprintf(`"%s" AS "%s"`, column.Source.Field, column.Alias)
	}

	return fmt.Sprintf(`"%s"."%s" AS "%s"`, column.Source.Alias, column.Source.Field, column.Alias)
}

func (_ *Postgres) SerializeColumn(column queries.Column) string {
	return fmt.Sprintf(`"%s"`, column.Source.Field)
}

func (_ *Postgres) SerializeTableSource(table queries.TableSource) string {
	schema := table.Schema

	if schema != "" {
		schema = fmt.Sprintf(`"%s".`, schema)
	}

	if table.Alias == "" {
		return fmt.Sprintf(`%s"%s"`, schema, table.Table)
	}

	return fmt.Sprintf(`%s"%s" AS "%s"`, schema, table.Table, table.Alias)
}

func (_ *Postgres) SerializeColumnKey(key queries.ColumnValue) string {
	if key.Alias == "" {
		return fmt.Sprintf(`"%s"`, key.Field)
	}

	return fmt.Sprintf(`"%s"."%s"`, key.Alias, key.Field)
}

func (_ *Postgres) SerializeAlias(raw string, alias string) string {
	return fmt.Sprintf(`%s AS "%s"`, raw, alias)
}
