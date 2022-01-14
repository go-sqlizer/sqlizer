package drivers

import (
	"database/sql"
	"fmt"
	"github.com/go-sqlizer/sqlizer/queries"
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
		whereOperators:    whereOperators,
		joinOperators:     joinOperators,
		orderOperators:    orderOperators,
		functionOperators: functionOperators,
	}

	return nil
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

	return fmt.Sprintf(`"%s"."%s" AS "%s"`, schema, table.Table, table.Alias)
}

func (_ Postgres) SerializeColumnKey(key queries.ColumnKey) string {
	return fmt.Sprintf(`"%s"."%s"`, key.Alias, key.Field)
}

func (_ Postgres) SerializeAlias(raw string, alias string) string {
	return fmt.Sprintf(`%s AS "%s"`, raw, alias)
}
