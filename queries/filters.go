package queries


func Or(where ...Where) Where {
	return Where{Nested: where, Operator: "or"}
}

func And(where ...Where) Where {
	return Where{Nested: where, Operator: "and"}
}

func Not(where ...Where) Where {
	return Where{Nested: where, Operator: "!"}
}

func Eq(key ColumnKey, value interface{}) Where {
	return Where{Key: key, Value: value, Operator: "="}
}

func NotEq(key ColumnKey, value interface{}) Where {
	return Where{Key: key, Value: value, Operator: "!="}
}

func IsNull(key ColumnKey) Where {
	return Where{Key: key, Operator: "null"}
}

func IsNotNull(key ColumnKey) Where {
	return Where{Key: key, Operator: "!null"}
}

func In(key ColumnKey, value []interface{}) Where {
	return Where{Key: key, Value: value, Operator: "in"}
}

func NotIn(key ColumnKey, value []interface{}) Where {
	return Where{Key: key, Value: value, Operator: "!in"}
}
