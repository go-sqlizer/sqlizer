package drivers

import (
	"database/sql"
	"fmt"
	"github.com/go-sqlizer/sqlizer/queries"
	"github.com/go-sqlizer/sqlizer/types"
	"strings"
)

type Driver interface {
	Connect(Config) error
	Select(query queries.BasicQuery) (*sql.Rows, error)
	Insert(insert queries.BasicQuery) (sql.Result, error)
	InsertReturning(insert queries.BasicQuery) *sql.Row
	Update(insert queries.BasicQuery) (sql.Result, error)
	UpdateReturning(insert queries.BasicQuery) *sql.Row
	Delete(delete queries.BasicQuery) (sql.Result, error)
	Transaction(func(Transaction) error) error
	Close()
}

type CommonDriver struct {
	db                *sql.DB
	serializer        queries.SQLSerializer
	WhereOperators    map[string]WhereOperator
	JoinOperators     map[queries.JoinType]JoinOperator
	OrderOperators    map[queries.OrderType]OrderOperator
	FunctionOperators map[string]FunctionOperator
}

type Transaction types.Transaction

func (driver *CommonDriver) Select(query queries.BasicQuery) (*sql.Rows, error) {
	var extra []string
	seq := valueSequence()
	columns, values := driver.renderColumns(&query.Columns, "", seq)
	joins, newValues := driver.renderJoins(query.Joins, seq)
	values = append(values, newValues...)

	if len(query.Where) > 0 {
		where, newValues := driver.WhereOperators["and"](nil, query.Where, driver, seq)
		extra = append(extra, fmt.Sprintf("WHERE %s", where))
		values = append(values, newValues...)
	}

	if len(query.Group) > 0 {
		group := driver.renderGroups(query.Group)
		extra = append(extra, fmt.Sprintf("GROUP BY %s", group))
	}

	if len(query.Order) > 0 {
		order := driver.renderOrder(query.Order)
		extra = append(extra, fmt.Sprintf("ORDER BY %s", order))
	}

	if query.Limit != nil {
		extra = append(extra, fmt.Sprintf("LIMIT %driver", *query.Limit))
	}

	if query.Offset != nil {
		extra = append(extra, fmt.Sprintf("OFFSET %driver", *query.Offset))
	}

	statement := fmt.Sprintf(
		"SELECT %s \nFROM %s\n%s\n%s",
		strings.Join(columns, ",\n\t"),
		driver.serializer.SerializeTableSource(query.From),
		strings.Join(joins, "\n"),
		strings.Join(extra, "\n"),
	)

	if query.Logging != nil {
		fmt.Println(statement, values)
	}

	if query.Transaction != nil {
		return query.Transaction.Query(statement, values...)
	} else {
		return driver.db.Query(statement, values...)
	}
}

func (driver *CommonDriver) Insert(insert queries.BasicQuery) (sql.Result, error) {
	var columns []string
	var valueSeq []string
	var values []interface{}
	seq := valueSequence()

	for _, column := range insert.Columns {
		columns = append(columns, driver.serializer.SerializeColumn(column))
		valueSeq = append(valueSeq, seq())
		values = append(values, column.Value)
	}

	statement := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s);",
		driver.serializer.SerializeTableSource(insert.From),
		strings.Join(columns, ", "),
		strings.Join(valueSeq, ", "),
	)

	if insert.Logging != nil {
		fmt.Println(statement, values)
	}

	if insert.Transaction != nil {
		return insert.Transaction.Exec(statement, values...)
	} else {
		return driver.db.Exec(statement, values...)
	}
}

func (driver *CommonDriver) Update(update queries.BasicQuery) (sql.Result, error) {
	var columns []string
	var values []interface{}
	var extra []string
	seq := valueSequence()

	for _, column := range update.Columns {
		if column.Value != nil {
			columns = append(columns, fmt.Sprintf("%s = %s", driver.serializer.SerializeColumn(column), seq()))
			values = append(values, column.Value)
		}
	}

	if len(update.Where) > 0 {
		where, newValues := driver.WhereOperators["and"](nil, update.Where, driver, seq)
		extra = append(extra, fmt.Sprintf("WHERE %s", where))
		values = append(values, newValues...)
	}

	statement := fmt.Sprintf(
		"UPDATE %s SET %s %s;",
		driver.serializer.SerializeTableSource(update.From),
		strings.Join(columns, ", "),
		strings.Join(extra, "\n"),
	)

	if update.Logging != nil {
		fmt.Println(statement, values)
	}

	if update.Transaction != nil {
		return update.Transaction.Exec(statement, values...)
	} else {
		return driver.db.Exec(statement, values...)
	}
}

func (driver *CommonDriver) Delete(update queries.BasicQuery) (sql.Result, error) {
	var values []interface{}
	var extra []string
	seq := valueSequence()

	if len(update.Where) > 0 {
		where, newValues := driver.WhereOperators["and"](nil, update.Where, driver, seq)
		extra = append(extra, fmt.Sprintf("WHERE %s", where))
		values = append(values, newValues...)
	}

	statement := fmt.Sprintf(
		"DELETE FROM %s %s;",
		driver.serializer.SerializeTableSource(update.From),
		strings.Join(extra, "\n"),
	)

	if update.Logging != nil {
		fmt.Println(statement, values)
	}

	if update.Transaction != nil {
		return update.Transaction.Exec(statement, values...)
	} else {
		return driver.db.Exec(statement, values...)
	}
}

func (driver *CommonDriver) Transaction(callback func(Transaction) error) (err error) {
	tx, err := driver.db.Begin()
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = callback(tx)
	return err
}

func (driver *CommonDriver) Close() {
	err := driver.db.Close()
	if err != nil {
		panic(err)
	}
}

type ValueSequencer func() string

func valueSequence() ValueSequencer {
	num := 0
	return func() string {
		num += 1
		return fmt.Sprintf("$%d", num)
	}
}

func (driver *CommonDriver) renderColumns(columns *[]queries.Column, prefix string, seq ValueSequencer) ([]string, []interface{}) {
	var strColumns []string
	var values []interface{}

	for _, column := range *columns {
		var columnAlias string
		if prefix == "" {
			columnAlias = column.Alias
		} else {
			columnAlias = fmt.Sprintf("%s.%s", prefix, column.Alias)
		}

		if column.Nested != nil {
			c, v := driver.renderColumns(column.Nested, columnAlias, seq)
			strColumns = append(strColumns, c...)
			values = append(values, v...)
		} else if column.Source != nil {
			column.Alias = columnAlias
			strColumns = append(strColumns, driver.serializer.SerializeColumnAlias(column))
		} else if column.Function != nil {
			c, v := driver.FunctionOperators[column.Function.Operator](column.Function.Values, driver, seq)
			strColumns = append(strColumns, driver.serializer.SerializeAlias(c, columnAlias))
			values = append(values, v...)
		}

	}

	return strColumns, values
}

type WhereOperator func(queries.SQLRender, interface{}, *CommonDriver, ValueSequencer) (string, []interface{})

func whereComparators(operator string) WhereOperator {
	return func(key queries.SQLRender, value interface{}, driver *CommonDriver, seq ValueSequencer) (filter string, values []interface{}) {
		var valueStr string
		if value == nil {
			filter = key.ToSQL(driver.serializer)
		} else {
			switch value.(type) {
			case queries.ColumnValue:
				valueStr = driver.serializer.SerializeColumnKey(value.(queries.ColumnValue))
			default:
				values = []interface{}{value}
				valueStr = seq()
			}

			filter = fmt.Sprintf("%s %s %s", key.ToSQL(driver.serializer), operator, valueStr)
		}

		return
	}
}

func whereNested(linker string) WhereOperator {
	return func(_ queries.SQLRender, value interface{}, driver *CommonDriver, seq ValueSequencer) (filter string, values []interface{}) {
		wheres := value.([]queries.Where)

		if len(wheres) == 0 {
			filter = ""
			return
		}

		var filters []string
		for _, where := range wheres {
			f, v := driver.WhereOperators[where.Operator](where.Key, where.Value, driver, seq)
			if len(f) > 0 {
				filters = append(filters, f)
				values = append(values, v...)
			}
		}

		filter = fmt.Sprintf("(%s)", strings.Join(filters, linker))
		return
	}
}

var whereOperators = map[string]WhereOperator{
	"and": whereNested(" AND "),
	"or":  whereNested(" OR "),
	"not": whereNested(" NOT "),
	"=":   whereComparators("="),
	"!=":  whereComparators("!="),
	">":   whereComparators(">"),
	">=":  whereComparators(">="),
	"<":   whereComparators("<"),
	"<=":  whereComparators("<="),
	"in":  nil,
}

type JoinOperator func(queries.Join, *CommonDriver, ValueSequencer) (string, []interface{})

func commonJoin(joinStr string) JoinOperator {
	return func(join queries.Join, driver *CommonDriver, seq ValueSequencer) (string, []interface{}) {
		where, values := driver.WhereOperators["and"](nil, join.Where, driver, seq)
		return fmt.Sprintf(`%s %s ON %s`, joinStr, driver.serializer.SerializeTableSource(join.To), where), values
	}
}

var joinOperators = map[queries.JoinType]JoinOperator{
	queries.InnerJoin: commonJoin("INNER JOIN"),
	queries.LeftJoin:  commonJoin("LEFT JOIN"),
	queries.RightJoin: commonJoin("RIGHT JOIN"),
}

func (driver *CommonDriver) renderJoins(joins []queries.Join, seq ValueSequencer) ([]string, []interface{}) {
	var strJoins []string
	var values []interface{}

	for _, join := range joins {
		j, v := driver.JoinOperators[join.Type](join, driver, seq)
		strJoins = append(strJoins, j)
		values = append(values, v...)
	}

	return strJoins, values
}

func (driver *CommonDriver) renderGroups(groups []queries.ColumnValue) string {
	var groupStr []string

	for _, group := range groups {
		groupStr = append(groupStr, driver.serializer.SerializeColumnKey(group))
	}

	return strings.Join(groupStr, ", ")
}

type OrderOperator func(queries.SQLRender, *CommonDriver) string

func commonOrder(order string) OrderOperator {
	return func(col queries.SQLRender, driver *CommonDriver) string {
		return fmt.Sprintf(`%s %s`, col.ToSQL(driver.serializer), order)
	}
}

var orderOperators = map[queries.OrderType]OrderOperator{
	queries.DescOrder: commonOrder("DESC"),
	queries.AscOrder:  commonOrder("ASC"),
}

func (driver *CommonDriver) renderOrder(orders []queries.Order) string {
	var orderStr []string

	for _, order := range orders {
		orderStr = append(orderStr, driver.OrderOperators[order.Type](order.Key, driver))
	}

	return strings.Join(orderStr, ", ")
}

type FunctionOperator func([]interface{}, *CommonDriver, ValueSequencer) (string, []interface{})

func commonFunction(fnName string, extra string) FunctionOperator {
	return func(fnValues []interface{}, driver *CommonDriver, seq ValueSequencer) (string, []interface{}) {
		var valueStr []string
		var values []interface{}

		for _, value := range fnValues {
			switch value.(type) {
			case queries.ColumnValue:
				valueStr = append(valueStr, driver.serializer.SerializeColumnKey(value.(queries.ColumnValue)))
			default:
				values = append(values, value)
				valueStr = append(valueStr, seq())
			}
		}

		return fmt.Sprintf(`%s(%s %s)`, fnName, extra, strings.Join(valueStr, ", ")), values
	}
}

var functionOperators = map[string]FunctionOperator{
	"count":     commonFunction("COUNT", ""),
	"max":       commonFunction("MAX", ""),
	"min":       commonFunction("MIN", ""),
	"avg":       commonFunction("AVG", ""),
	"countDist": commonFunction("COUNT", "DISTINCT"),
}
