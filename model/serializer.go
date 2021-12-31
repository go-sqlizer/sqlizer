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
		scanArgs, elem := generateValue(resultType, nil)

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

func SerializeResult() {

}

func generateValue(resultType reflect.Type, value *reflect.Value) ([]interface{}, reflect.Value) {
	var scanArgs []interface{}
	var elem reflect.Value
	if value != nil {
		elem = *value
	} else {
		elem = reflect.New(resultType).Elem()
	}

	switch resultType.Kind() {
	case reflect.Struct:
		for i := 0; i < elem.NumField(); i++ {
			elemField := elem.Field(i)

			switch elemField.Kind() {
			case reflect.Struct:
				newArgs, _ := generateValue(elemField.Type(), &elemField)
				scanArgs = append(scanArgs, newArgs...)
			case reflect.Ptr:
				insideElemPointer := reflect.New(elemField.Type().Elem())
				insideElem := insideElemPointer.Elem()
				newArgs, _ := generateValue(insideElem.Type(), &insideElem)
				elemField.Set(insideElemPointer)
				scanArgs = append(scanArgs, newArgs...)
			case reflect.Array, reflect.Slice:
				arrayType := elemField.Type().Elem()
				elemField.Set(reflect.Append(elemField, reflect.New(arrayType).Elem()))
				newElem := elemField.Index(0)
				newArgs, _ := generateValue(arrayType, &newElem)
				scanArgs = append(scanArgs, newArgs...)
			default:
				scanArgs = append(scanArgs, elemField.Addr().Interface())
			}
		}
	case reflect.Ptr:
		newArgs, _ := generateValue(elem.Type().Elem(), &elem)
		return newArgs, elem
	case reflect.Array, reflect.Slice:
		newArgs, newValue := generateValue(elem.Type().Elem(), nil)
		elem.Set(reflect.Append(elem, newValue))
		return newArgs, newValue
	default:
		return []interface{}{value.Addr().Interface()}, reflect.Value{}
	}

	return scanArgs, elem
}
