package queries

type SQLSerializer interface {
	SerializeColumnKey(key ColumnKey) string
	SerializeTableSource(source TableSource) string
	SerializeColumn(source Column) string
	SerializeAlias(raw string, alias string) string
}

type SQLRender interface {
	ToSQL(SQLSerializer) string
}
