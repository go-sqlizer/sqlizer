package model

import (
	"errors"
	"fmt"
	"github.com/Supersonido/sqlizer/queries"
	"reflect"
)

func (model Model) FindAll(result interface{}, options queries.Options) error {
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
		fmt.Println(err)
		return err
	}

	return SerializeResults(resultListValue, query, rows)
}

func (model Model) Paginate(result interface{}, options queries.Options) error {
	return nil
}
