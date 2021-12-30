package drivers

import (
	"database/sql"
	"fmt"
)

type Postgres struct {
	driver
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

func (_ Postgres) SelectColumn(tableAlias string, field string, modelAlias string, fieldAlias string) string {
	if modelAlias == "" {
		return fmt.Sprintf(`"%s"."%s" AS "%s"`, tableAlias, field, fieldAlias)
	}

	return fmt.Sprintf(`"%s"."%s" AS "%s.%s"`, tableAlias, field, modelAlias, fieldAlias)
}

func (_ Postgres) Column(tableAlias string, field string) string {
	return fmt.Sprintf(`"%s"."%s"`, tableAlias, field)
}

func (_ Postgres) From(schema string, table string, alias string) string {
	if schema == "" {
		schema = "public"
	}

	return fmt.Sprintf(`"%s"."%s" AS "%s"`, schema, table, alias)
}

func (p Postgres) Join(jType JoinType, schema string, table string, childAlias string, childField string, parentAlias string, parentField string) string {

	return fmt.Sprintf(
		`%s %s ON "%s"."%s" = "%s"."%s"`,
		jType,
		p.From(schema, table, childAlias),
		childAlias, childField,
		parentAlias, parentField,
	)
}

func (p Postgres) Operator(action string) string {
	return ""
}

func (p Postgres) Close() {
	return
}

func (p Postgres) connectionString(config Config) string {
	return fmt.Sprintf(
		`host=%s port=%d user=%s password="%s" dbname=%s sslmode=%s`,
		config.Host, config.Port, config.User, config.Password, config.Name, config.SSl)

}

func (p Postgres) Exec(query Query) error {
	value, err := p.db.Exec(query.Statement, query.Values...)
	fmt.Println(value)
	if err != nil {
		return err
	}
	return nil
}

func (p Postgres) Query(query Query) (*sql.Rows, error) {
	return p.db.Query(query.Statement, query.Values...)
}
