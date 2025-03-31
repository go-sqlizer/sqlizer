package model

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/go-sqlizer/sqlizer/common"
	"github.com/go-sqlizer/sqlizer/queries"
	"reflect"
	"time"
)

type rowHashTable struct {
	NestedHash *map[string]rowHashTable
	Elem       *reflect.Value
	Index      int
}

type SqlRows interface {
	Err() error
	Scan(dest ...interface{}) error
	Next() bool
}

type SqlRow interface {
	Scan(dest ...interface{}) error
	Err() error
}

func SerializeResults(result reflect.Value, query queries.BasicQuery, row SqlRows) (err error) {
	defer common.CaptureError(&err, "Invalid Destination Model")

	err = row.Err()
	if err != nil {
		return err
	}

	// ResultInformation
	resultListType := result.Type()
	resultAux := reflect.MakeSlice(resultListType, 0, 0)
	resultHashTable := make(map[string]rowHashTable)

	// Generate basic row result
	resultType := resultListType.Elem()
	for row.Next() {
		scanArgs, resultArgs, argsStruct := generateValues(query.Columns)
		if err = row.Scan(scanArgs...); err != nil {
			return err
		}

		runColumnSetters(scanArgs, resultArgs, query.Columns)
		processNewValue(&query, &resultAux, &resultType, &argsStruct, &resultHashTable)
	}

	result.Set(resultAux)
	return nil
}

func SerializeResult(result reflect.Value, query queries.BasicQuery, row SqlRow) (err error) {
	defer common.CaptureError(&err, "Invalid Destination Model")

	err = row.Err()
	if err != nil {
		return err
	}

	// Generate basic row result
	scanArgs, resultArgs, argsStruct := generateValues(query.Columns)
	if err = row.Scan(scanArgs...); err != nil {
		return err
	}

	runColumnSetters(scanArgs, resultArgs, query.Columns)
	setValues(&result, &argsStruct, query.Columns, nil, "")
	return nil
}

func generateValues(columns []queries.Column) ([]interface{}, []interface{}, map[string]interface{}) {
	var scanArgs []interface{}
	var resultArgs []interface{}

	argsStruct := make(map[string]interface{})

	for _, column := range columns {
		if column.Nested == nil {
			// Render columnArgs
			columnType := *column.ColumnType
			columnInstance := reflect.New(columnType)
			if columnType.Kind() == reflect.Ptr {
				scanArgs = append(scanArgs, columnInstance.Interface())
			} else {
				newTest := reflect.New(columnInstance.Type())
				newTest.Elem().Set(columnInstance)
				scanArgs = append(scanArgs, newTest.Interface())
			}

			// Render resultArgs
			valueType := *column.Type
			valueInstance := reflect.New(valueType)
			if valueType.Kind() == reflect.Ptr {
				argsStruct[column.Alias] = valueInstance
				resultArgs = append(resultArgs, valueInstance.Interface())
			} else {
				newTest := reflect.New(valueInstance.Type())
				newTest.Elem().Set(valueInstance)
				argsStruct[column.Alias] = newTest.Elem()
				resultArgs = append(resultArgs, newTest.Interface())
			}
		} else {
			newScanArgs, newResultArgs, newNestedValues := generateValues(*column.Nested)
			scanArgs = append(scanArgs, newScanArgs...)
			resultArgs = append(resultArgs, newResultArgs...)
			argsStruct[column.Alias] = newNestedValues
		}
	}

	return scanArgs, resultArgs, argsStruct
}

func runColumnSetters(scanArgs []interface{}, resultArgs []interface{}, columns []queries.Column) {
	for _, column := range columns {
		if column.Nested == nil {
			scanArg, scanArgsTmp := scanArgs[0], scanArgs[1:]
			resultArg, resultArgsTmp := resultArgs[0], resultArgs[1:]
			scanArgs = scanArgsTmp
			resultArgs = resultArgsTmp

			if column.Get != nil {
				value := reflect.ValueOf(scanArg)
				actualElem := value.Elem()

				result := reflect.ValueOf(resultArg)
				actualResult := result.Elem()

				getResult := column.Get(actualElem.Interface())
				if getResult != nil {
					getResultValue := reflect.ValueOf(getResult)
					actualResult.Set(getResultValue)
				}
			} else {
				value := reflect.ValueOf(scanArg).Elem()
				result := reflect.ValueOf(resultArg).Elem()

				result.Set(value)
			}

		} else {
			runColumnSetters(scanArgs, resultArgs, *column.Nested)
		}
	}
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
	default:
		panic("Invalid return type " + resultType.Name())
	}
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
		default:
			continue
		}
	}

	return elem
}

func renderValuePtr(elem reflect.Value) reflect.Value {
	// @TODO Be able to handle internal types better
	if elem.Type() == reflect.TypeOf((*time.Time)(nil)) {
		return elem
	}

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
	default:
		return newElem
	}
}

func renderValueSlice(elem reflect.Value) reflect.Value {
	return elem
}

func processNewValue(query *queries.BasicQuery, result *reflect.Value, resultType *reflect.Type, row *map[string]interface{}, resultHashTable *map[string]rowHashTable) {
	var resultInstance *reflect.Value
	var nestedHashTable *map[string]rowHashTable

	valueHash := rowHash("", query.Columns, row)
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
			Index:      result.Len() - 1,
		}
	}

	setValues(resultInstance, row, query.Columns, nestedHashTable, query.From.Alias)
	result.Index((*resultHashTable)[valueHash].Index).Set(*resultInstance)
}

func setValues(result *reflect.Value, row *map[string]interface{}, columns []queries.Column, resultHashTable *map[string]rowHashTable, prefix string) (length uint, nilCounter uint) {
	result = common.ValueFinder(result)

	if result.Kind() == reflect.Slice {
		// Don't render row when primaryKey is null
		if rowNil(columns, row) {
			result.Set(reflect.MakeSlice(result.Type(), 0, 0))
			return
		}

		valueHash := rowHash(prefix, columns, row)
		resultInstanceAux, ok := (*resultHashTable)[valueHash]
		nilCounter, length = 0, 0

		if ok {
			resultInstance := resultInstanceAux.Elem
			nestedHashTable := resultInstanceAux.NestedHash
			setValues(resultInstance, row, columns, nestedHashTable, prefix)
			result.Index(resultInstanceAux.Index).Set(*resultInstance)
		} else {
			newVal := renderValue(result.Type().Elem())
			newHash := make(map[string]rowHashTable)

			n, nc := setValues(&newVal, row, columns, &newHash, prefix)
			if n > 0 && n != nc {
				result.Set(reflect.Append(*result, newVal))
				newVal = result.Index(result.Len() - 1)

				(*resultHashTable)[valueHash] = rowHashTable{
					NestedHash: &newHash,
					Elem:       &newVal,
					Index:      result.Len() - 1,
				}
			}

		}

		return
	}

	for _, column := range columns {
		fieldName := column.Alias
		item := (*row)[fieldName]
		length++

		switch item.(type) {
		case reflect.Value:
			val := item.(reflect.Value)
			if !val.IsNil() {
				if resultField := result.FieldByName(fieldName); resultField.IsValid() {
					resultField.Set(val.Elem())
				}
			} else {
				nilCounter++
			}
		case map[string]interface{}:
			nestedValue := result.FieldByName(fieldName)
			nestedRow := item.(map[string]interface{})

			if n, nc := setValues(&nestedValue, &nestedRow, *column.Nested, resultHashTable, fieldName); n > 0 && n == nc {
				nestedValue.Set(reflect.Zero(nestedValue.Type()))
			}
		}
	}

	return
}

func rowHash(prefix string, columns []queries.Column, row *map[string]interface{}) string {
	strHash := prefix
	for _, column := range columns {
		if column.Type != nil && column.IsPrimaryKey {
			value := (*row)[column.Alias].(reflect.Value)
			hash := md5.Sum([]byte(value.Elem().String()))
			strHash += "#" + hex.EncodeToString(hash[:])
		}
	}

	return strHash
}

func rowNil(columns []queries.Column, row *map[string]interface{}) bool {
	for _, column := range columns {
		if column.Type != nil && column.IsPrimaryKey {
			value := (*row)[column.Alias].(reflect.Value)
			if value.IsNil() {
				return true
			}
		}
	}

	return false
}
