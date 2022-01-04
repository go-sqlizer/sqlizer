package model

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/Supersonido/sqlizer/queries"
	"reflect"
)

type rowHashTable struct {
	NestedHash *map[string]rowHashTable
	Elem       *reflect.Value
}

func SerializeResults(result reflect.Value, query queries.SelectQuery, row *sql.Rows) error {
	err := row.Err()
	if row.Err() != nil {
		fmt.Println(err)
		return row.Err()
	}

	// ResultInformation
	resultListType := result.Type()
	resultAux := reflect.MakeSlice(resultListType, 0, 0)
	resultHashTable := make(map[string]rowHashTable)

	// Generate basic row result
	resultType := resultListType.Elem()

	for row.Next() {
		scanArgs, argsStruct := generateValues(query.Columns)

		if err = row.Scan(scanArgs...); err != nil {
			fmt.Println(err)
			return err
		}

		if err = processNewValue(&query, &resultAux, &resultType, &argsStruct, &resultHashTable); err != nil {
			fmt.Println(err)
			return err
		}
	}

	result.Set(resultAux)
	return nil
}

func generateValues(columns []queries.Column) ([]interface{}, map[string]interface{}) {
	var scanArgs []interface{}
	argsStruct := make(map[string]interface{})

	for _, column := range columns {
		if column.Type != nil {
			valueType := *column.Type
			valueInstance := reflect.New(valueType)
			if valueType.Kind() == reflect.Ptr {
				argsStruct[column.Alias] = valueInstance
				scanArgs = append(scanArgs, valueInstance.Interface())
			} else {
				newTest := reflect.New(valueInstance.Type())
				newTest.Elem().Set(valueInstance)
				argsStruct[column.Alias] = newTest.Elem()
				scanArgs = append(scanArgs, newTest.Interface())
			}
		} else {
			newScanArgs, newNestedValues := generateValues(column.Nested)
			scanArgs = append(scanArgs, newScanArgs...)
			argsStruct[column.Alias] = newNestedValues
		}
	}

	return scanArgs, argsStruct
}

func renderValue(resultType reflect.Type) reflect.Value {
	switch resultType.Kind() {
	case reflect.Struct:
		elem := reflect.New(resultType).Elem()
		return renderValueStruct(elem)
	case reflect.Ptr:
		elem := reflect.New(resultType).Elem()
		return renderValuePtr(elem)
	case reflect.Array, reflect.Slice:
		elem := reflect.MakeSlice(resultType, 0, 0).Elem()
		return renderValueSlice(elem)
	}

	panic("Invalid return type " + resultType.Name())
}

func renderValueStruct(elem reflect.Value) reflect.Value {
	for i := 0; i < elem.NumField(); i++ {
		elemField := elem.Field(i)

		switch elemField.Kind() {
		case reflect.Struct:
			_ = renderValueStruct(elemField)
		case reflect.Ptr:
			newValue := renderValuePtr(elemField)
			elemField.Set(newValue)
		case reflect.Array, reflect.Slice:
		}
	}

	return elem
}

func renderValuePtr(elem reflect.Value) reflect.Value {
	var newElem reflect.Value
	if elem.IsZero() || elem.IsNil() {
		newElem = reflect.New(elem.Type().Elem())
	} else {
		newElem = reflect.New(elem.Elem().Type()).Elem()
	}

	switch newElem.Kind() {
	case reflect.Struct:
		return renderValueStruct(newElem)
	case reflect.Ptr:
		_ = renderValuePtr(newElem)
		elem.Set(newElem)
		return newElem
	case reflect.Array, reflect.Slice:
		return renderValueSlice(newElem)
	}

	return newElem
}

func renderValueSlice(elem reflect.Value) reflect.Value {
	return elem
}

func processNewValue(query *queries.SelectQuery, result *reflect.Value, resultType *reflect.Type, row *map[string]interface{}, resultHashTable *map[string]rowHashTable) error {
	var resultInstance *reflect.Value
	var nestedHashTable *map[string]rowHashTable

	valueHash := rowHash(query.Columns, row)
	if resultInstanceAux, ok := (*resultHashTable)[valueHash]; ok {
		resultInstance = resultInstanceAux.Elem
		nestedHashTable = resultInstanceAux.NestedHash
	} else {
		// Create new Value
		*result = reflect.Append(*result, renderValue(*resultType))
		newVal := result.Index(result.Len() - 1)
		resultInstance = &newVal

		// Create nested hash table
		newHash := make(map[string]rowHashTable)
		nestedHashTable = &newHash

		(*resultHashTable)[valueHash] = rowHashTable{
			NestedHash: nestedHashTable,
			Elem:       &newVal,
		}
	}

	setValues(resultInstance, row, query.Columns, nestedHashTable)
	return nil
}

func setValues(result *reflect.Value, row *map[string]interface{}, columns []queries.Column, resultHashTable *map[string]rowHashTable) (length uint, nilCounter uint) {
	result = valueFinder(result)

	if result.Kind() == reflect.Slice {
		var resultInstance *reflect.Value
		var nestedHashTable *map[string]rowHashTable

		valueHash := rowHash(columns, row)
		resultInstanceAux, ok := (*resultHashTable)[valueHash]
		if ok {
			resultInstance = resultInstanceAux.Elem
			nestedHashTable = resultInstanceAux.NestedHash
		} else {
			// Create new Value
			newVal := reflect.New(result.Type().Elem()).Elem()
			resultInstance = &newVal

			// Create nested hash table
			newHash := make(map[string]rowHashTable)
			nestedHashTable = &newHash

			(*resultHashTable)[valueHash] = rowHashTable{
				NestedHash: nestedHashTable,
				Elem:       &newVal,
			}
		}

		n, nc := setValues(resultInstance, row, columns, nestedHashTable)
		nilCounter, length = 0, 0
		if !ok && n > 0 && n != nc {
			result.Set(reflect.Append(*result, *resultInstance))
		}
		return
	}

	for _, column := range columns {
		fieldName := column.Alias
		item := (*row)[fieldName]
		length++

		switch item.(type) {
		case reflect.Value:
			if !item.(reflect.Value).IsNil() {
				result.FieldByName(fieldName).Set(item.(reflect.Value).Elem())
			} else {
				nilCounter++
			}
		case map[string]interface{}:
			nestedValue := result.FieldByName(fieldName)
			nestedRow := item.(map[string]interface{})

			if n, nc := setValues(&nestedValue, &nestedRow, column.Nested, resultHashTable); n > 0 && n == nc {
				nestedValue.Set(reflect.Zero(nestedValue.Type()))
			}
		}
	}

	return
}

func valueFinder(result *reflect.Value) *reflect.Value {
	switch result.Kind() {
	case reflect.Ptr:
		newResult := result.Elem()
		return &newResult
	}

	return result
}

func rowHash(columns []queries.Column, row *map[string]interface{}) string {
	strHash := ""
	for _, column := range columns {
		if column.Type != nil && column.IsPrimaryKey {
			value := (*row)[column.Alias].(reflect.Value)
			hash := md5.Sum([]byte(value.Elem().String()))
			strHash += hex.EncodeToString(hash[:]) + "#"
		}
	}

	return strHash
}
