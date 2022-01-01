package model

import (
	"database/sql"
	"fmt"
	"github.com/Supersonido/sqlizer/drivers"
	"reflect"
)

func SerializeResults(result interface{}, query drivers.Query, row *sql.Rows) error {
	if row.Err() != nil {
		return row.Err()
	}

	resultListType := reflect.TypeOf(result)
	resultType := resultListType.Elem().Elem()
	resultAux := reflect.MakeSlice(resultListType.Elem(), 0, 1)

	for row.Next() {
		scanArgs, elem := generateValue(resultType)

		// End of rows
		if err := row.Scan(scanArgs...); err != nil {
			fmt.Println(err)
			return err
		}

		resultAux = reflect.Append(resultAux, elem)
	}

	reflect.ValueOf(result).Elem().Set(resultAux)
	return nil
}

func generateValue(resultType reflect.Type) ([]interface{}, reflect.Value) {
	switch resultType.Kind() {
	case reflect.Struct:
		elem := reflect.New(resultType).Elem()
		return generateValueStruct(elem)
	case reflect.Ptr:
	case reflect.Array, reflect.Slice:
	}

	panic("Invalid return type " + resultType.Name())
}

func generateValueStruct(elem reflect.Value) ([]interface{}, reflect.Value) {
	var scanArgs []interface{}

	for i := 0; i < elem.NumField(); i++ {
		elemField := elem.Field(i)

		switch elemField.Kind() {
		case reflect.Struct:
			newArgs, _ := generateValueStruct(elemField)
			scanArgs = append(scanArgs, newArgs...)
		case reflect.Ptr:
			newArgs, newValue := generateValuePtr(elemField)
			elemField.Set(newValue)
			if len(newArgs) == 0 {
				newArgs = []interface{}{elemField.Addr().Interface()}
			}

			scanArgs = append(scanArgs, newArgs...)
		case reflect.Array, reflect.Slice:
		default:
			scanArgs = append(scanArgs, elemField.Addr().Interface())
		}
	}

	return scanArgs, elem
}

func generateValuePtr(elem reflect.Value) ([]interface{}, reflect.Value) {
	var newElem reflect.Value
	if elem.IsZero() || elem.IsNil() {
		newElem = reflect.New(elem.Type().Elem())
	} else {
		newElem = reflect.New(elem.Elem().Type()).Elem()
	}

	switch newElem.Kind() {
	case reflect.Struct:
		return generateValueStruct(newElem)
	case reflect.Ptr:
		scanArgs, _ := generateValuePtr(newElem)
		elem.Set(newElem)
		return scanArgs, newElem
	case reflect.Array, reflect.Slice:
	default:
		return []interface{}{}, newElem
	}

	panic("")
}
