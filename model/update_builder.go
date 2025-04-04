package model

import (
	"github.com/go-sqlizer/sqlizer/queries"
	"reflect"
)

func UpdateBuilder(data reflect.Value, result *reflect.Type, model *Model, options queries.UpdateOptions) queries.BasicQuery {
	var columns []queries.Column

	modelColumns := reflect.ValueOf(model.Columns)
	modelColumnsType := modelColumns.Type()
	for i := 0; i < modelColumns.NumField(); i++ {
		modelFieldValue := modelColumns.Field(i)
		fieldName := modelColumnsType.Field(i).Name
		field := modelFieldValue.Interface().(Field)
		columnType := reflect.Type(field.Type)
		fieldType := reflect.Type(field.Type)
		dataValue := data.FieldByName(fieldName)

		column := queries.Column{
			Alias:      fieldName,
			ColumnType: &columnType,
			Source: &queries.ColumnSource{
				Field: field.Field,
			},
			IsPrimaryKey: field.PrimaryKey,
		}

		if model.Timestamps != nil && model.Timestamps.UpdatedAt != nil && model.Timestamps.UpdatedAt.Field == fieldName {
			columns = append(columns, renderTimestamp(model.Timestamps.UpdatedAt, modelColumns))
		}

		if dataValue.Kind() == reflect.Ptr && dataValue.IsNil() && field.DefaultValue != nil {
			if v := reflect.ValueOf(field.DefaultValue); v.Kind() == reflect.Func {
				column.Value = v.Call([]reflect.Value{})[0].Interface()
			} else {
				column.Value = field.DefaultValue
			}
		} else if dataValue.IsValid() {
			column.Value = dataValue.Interface()
		}

		if field.Set != nil {
			column.Value = field.Set(column.Value)
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
			Where:       options.Where,
		},
		Columns: columns,
		From: queries.TableSource{
			Schema: model.Schema,
			Table:  model.Table,
		},
	}
}
