package model

import (
	"errors"
	"github.com/go-sqlizer/sqlizer/queries"
	"reflect"
)

func (model Model) FindAll(result interface{}, options queries.QueryOptions) error {
	resultPointerListValue := reflect.ValueOf(result)
	if resultPointerListValue.Kind() != reflect.Ptr {
		return errors.New("result must start as a pointer")
	}

	resultListValue := resultPointerListValue.Elem()
	if resultListValue.Kind() != reflect.Slice {
		return errors.New("result must be a slice")
	}

	resultType := resultListValue.Type().Elem()

	query := SelectBuilder(resultType, model, options)
	rows, err := model.driver.Select(query)

	if err != nil {
		return err
	}

	return SerializeResults(resultListValue, query, rows)
}

//func (model Model) Paginate(result interface{}, options queries.QueryOptions) error {
//	return nil
//}

func (model Model) Insert(data interface{}, result interface{}, options queries.InsertOptions) error {
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() != reflect.Struct {
		return errors.New("the input data must be a struct")
	}

	if result != nil {
		resultPointerValue := reflect.ValueOf(result)
		if resultPointerValue.Kind() != reflect.Ptr {
			return errors.New("result must start as a pointer")
		}

		resultValue := resultPointerValue.Elem()
		if resultValue.Kind() != reflect.Struct {
			return errors.New("result must be a struct")
		}

		resultType := resultValue.Type()
		insert := InsertBuilder(dataValue, &resultType, model, options)
		row := model.driver.InsertReturning(insert)
		return SerializeResult(resultValue, insert, row)
	}

	insert := InsertBuilder(dataValue, nil, model, options)
	_, err := model.driver.Insert(insert)
	return err
}

//func (model Model) BulkInsert(data interface{}, result interface{}, options queries.InsertOptions) error {
//	return nil
//}

func (model Model) Update(data interface{}, result interface{}, options queries.UpdateOptions) error {
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() != reflect.Struct {
		return errors.New("the input data must be a struct")
	}

	if result != nil {
		resultPointerValue := reflect.ValueOf(result)
		if resultPointerValue.Kind() != reflect.Ptr {
			return errors.New("result must start as a pointer")
		}

		resultValue := resultPointerValue.Elem()
		if resultValue.Kind() != reflect.Struct {
			return errors.New("result must be a struct")
		}

		resultType := resultValue.Type()
		insert := UpdateBuilder(dataValue, &resultType, model, options)
		row := model.driver.UpdateReturning(insert)
		return SerializeResult(resultValue, insert, row)
	}

	insert := UpdateBuilder(dataValue, nil, model, options)
	_, err := model.driver.Update(insert)
	return err
}

func (model Model) UpdateByPk(pk interface{}, data interface{}, result interface{}, options queries.UpdateOptions) error {
	options.Where = []queries.Where{
		queries.Eq(queries.ColumnValue{Field: model.primaryKey.Field}, pk),
		queries.And(options.Where...),
	}

	return model.Update(data, result, options)
}
