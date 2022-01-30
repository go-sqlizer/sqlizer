package model

import (
	"github.com/go-sqlizer/sqlizer/queries"
	"reflect"
)

func InsertBuilder(data reflect.Value, result *reflect.Type, model Model, options queries.InsertOptions) queries.BasicQuery {
	var columns []queries.Column

	modelColumns := reflect.ValueOf(model.Columns)
	modelColumnsType := modelColumns.Type()
	for i := 0; i < modelColumns.NumField(); i++ {
		modelFieldValue := modelColumns.Field(i)
		fieldName := modelColumnsType.Field(i).Name
		field := modelFieldValue.Interface().(Field)
		fieldType := reflect.Type(field.Type)
		dataValue := data.FieldByName(fieldName)

		column := queries.Column{
			Alias: fieldName,
			Source: &queries.ColumnSource{
				Field: field.Field,
			},
			IsPrimaryKey: field.PrimaryKey,
		}

		if dataValue.IsValid() {
			column.Value = dataValue.Interface()
		} else if field.DefaultValue != nil {
			if v := reflect.ValueOf(field.DefaultValue); v.Kind() == reflect.Func {
				column.Value = v.Call([]reflect.Value{})[0].Interface()
			} else {
				column.Value = field.DefaultValue
			}
		}

		if result != nil {
			resultField, ok := (*result).FieldByName(fieldName)
			if ok {
				fieldType = resultField.Type
			}
		}

		column.Type = &fieldType
		columns = append(columns, column)
	}

	return queries.BasicQuery{
		QueryOptions: queries.QueryOptions{
			Logging:     options.Logging,
			Transaction: options.Transaction,
		},
		Columns: columns,
		From: queries.TableSource{
			Schema: model.Schema,
			Table:  model.Table,
		},
	}
}
