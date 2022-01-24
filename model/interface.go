package model

import (
	"errors"
	"github.com/go-sqlizer/sqlizer/queries"
	"math"
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

func (model Model) FindOne(result interface{}, options queries.QueryOptions) error {
	resultPointerValue := reflect.ValueOf(result)
	if resultPointerValue.Kind() != reflect.Ptr {
		return errors.New("result must start as a pointer")
	}

	resultValue := resultPointerValue.Elem()
	if resultValue.Kind() != reflect.Struct {
		return errors.New("result must be a struct")
	}

	resultType := resultValue.Type()
	query := SelectBuilder(resultType, model, options)
	rows, err := model.driver.Select(query)
	if err != nil {
		return err
	}

	rows.Next()
	return SerializeResult(resultValue, query, rows)
}

func (model Model) Count(options queries.QueryOptions) (*uint, error) {
	var count []struct{ Count uint }
	options.Fields = queries.Fields{
		Includes: []queries.Field{
			{As: "Count", Fn: queries.Count(queries.ColumnValue{Alias: model.Name, Field: model.primaryKey.Field})},
		},
	}

	if len(options.Group) > 0 {
		options.Fields.Includes = append(options.Fields.Includes, queries.Field{As: "Id"})
	}

	if err := model.FindAll(&count, options); err != nil {
		return nil, err
	}

	countLen := uint(len(count))
	if countLen == 1 {
		return &count[0].Count, nil
	}

	return &countLen, nil
}

func (model Model) Paginate(result interface{}, options queries.PaginateOptions) (*queries.PaginateResults, error) {
	total, err := model.Count(options.QueryOptions)
	if err != nil {
		return nil, err
	}

	pages := math.Ceil(float64(*total / options.PerPage))
	limit := int(options.PerPage)
	offset := int(options.PerPage * (options.Page - 1))

	options.Limit = &limit
	options.Offset = &offset
	if err = model.FindAll(result, options.QueryOptions); err != nil {
		return nil, err
	}

	return &queries.PaginateResults{Total: *total, Pages: uint(pages), Page: options.Page, PerPage: options.PerPage}, nil
}

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

func (model Model) Delete(options queries.DeleteOptions) error {
	_, err := model.driver.Delete(queries.BasicQuery{
		From: queries.TableSource{
			Schema: model.Schema,
			Table:  model.Table,
		},
		QueryOptions: queries.QueryOptions{
			Transaction: options.Transaction,
			Logging:     options.Logging,
			Where:       options.Where,
		},
	})
	return err
}

func (model Model) DeleteByPk(pk interface{}, options queries.DeleteOptions) error {
	options.Where = []queries.Where{
		queries.Eq(queries.ColumnValue{Field: model.primaryKey.Field}, pk),
		queries.And(options.Where...),
	}

	return model.Delete(options)
}
