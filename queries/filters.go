package queries

func Or(where ...Where) Where {
	return Where{Value: where, Operator: "or"}
}

func And(where ...Where) Where {
	return Where{Value: where, Operator: "and"}
}

func Not(where ...Where) Where {
	return Where{Value: where, Operator: "not"}
}

func Key(key SQLRender) Where {
	return Where{Key: key, Operator: "col"}
}

func Eq(key SQLRender, value interface{}) Where {
	return Where{Key: key, Value: value, Operator: "="}
}

func Gt(key SQLRender, value interface{}) Where {
	return Where{Key: key, Value: value, Operator: ">"}
}

func Gte(key SQLRender, value interface{}) Where {
	return Where{Key: key, Value: value, Operator: ">="}
}

func NotEq(key SQLRender, value interface{}) Where {
	return Where{Key: key, Value: value, Operator: "!="}
}

func IsNull(key SQLRender) Where {
	return Where{Key: key, Operator: "null"}
}

func IsNotNull(key SQLRender) Where {
	return Where{Key: key, Operator: "!null"}
}

func In(key SQLRender, value []interface{}) Where {
	return Where{Key: key, Value: value, Operator: "in"}
}

func NotIn(key SQLRender, value []interface{}) Where {
	return Where{Key: key, Value: value, Operator: "!in"}
}

func IsTrue(key SQLRender) Where {
	return Where{Key: key, Operator: "true"}
}

func IsFalse(key SQLRender) Where {
	return Where{Key: key, Operator: "false"}
}