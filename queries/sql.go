package queries

type SQLSerializer interface {
	SerializeColumnKey(key ColumnValue) string
	SerializeTableSource(source TableSource) string
	SerializeColumnAlias(source Column) string
	SerializeColumn(source Column) string
	SerializeAlias(raw string, alias string) string
}

type SQLRender interface {
	ToSQL(SQLSerializer) string
}
