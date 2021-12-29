package drivers

import "fmt"

type Postgres struct{}

func (p Postgres) Connect(config Config) error {
	return nil
}

func (_ Postgres) Column(tableAlias string, field string, modelAlias string, fieldAlias string) string {
	if modelAlias == "" {
		return fmt.Sprintf(`"%s"."%s" AS "%s"`, tableAlias, field, fieldAlias)
	}

	return fmt.Sprintf(`"%s"."%s" AS "%s.%s"`, tableAlias, field, modelAlias, fieldAlias)
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
