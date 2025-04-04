package model

import (
	"github.com/go-sqlizer/sqlizer/queries"
	"reflect"
)

func InsertBuilder(data reflect.Value, result *reflect.Type, model *Model, options queries.InsertOptions) queries.BasicQuery {
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
			Type:       &fieldType,
			ColumnType: &columnType,
			Source: &queries.ColumnSource{
				Field: field.Field,
			},
			IsPrimaryKey: field.PrimaryKey,
		}

		if model.Timestamps != nil {
			if model.Timestamps.CreatedAt != nil && model.Timestamps.CreatedAt.Field == fieldName {
				columns = append(columns, renderTimestamp(model.Timestamps.CreatedAt, modelColumns))
				continue
			}

			if model.Timestamps.UpdatedAt != nil && model.Timestamps.UpdatedAt.Field == fieldName {
				columns = append(columns, renderTimestamp(model.Timestamps.UpdatedAt, modelColumns))
				continue
			}
		}

		if dataValue.Kind() == reflect.Ptr && dataValue.IsNil() && field.DefaultValue != nil {
			if v := reflect.ValueOf(field.DefaultValue); v.Kind() == reflect.Func {
				column.Value = v.Call([]reflect.Value{})[0].Interface()
			} else {
				column.Value = field.DefaultValue
			}
		} else if dataValue.IsValid() {
			column.Value = dataValue.Interface()
		} else if field.DefaultValue != nil {
			if v := reflect.ValueOf(field.DefaultValue); v.Kind() == reflect.Func {
				column.Value = v.Call([]reflect.Value{})[0].Interface()
			} else {
				column.Value = field.DefaultValue
			}
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

func renderTimestamp(timestamp *Timestamp, modelColumns reflect.Value) queries.Column {
	createdAtValue := modelColumns.FieldByName(timestamp.Field)
	createdAt := createdAtValue.Interface().(Field)
	var value interface{}

	if timestamp.Value != nil {
		if v := reflect.ValueOf(timestamp.Value); v.Kind() == reflect.Func {
			value = v.Call([]reflect.Value{})[0].Interface()
		} else {
			value = timestamp.Value
		}
	}

	createdAtType := reflect.Type(createdAt.Type)
	return queries.Column{
		Alias:      timestamp.Field,
		ColumnType: &createdAtType,
		Type:       &createdAtType,
		Source: &queries.ColumnSource{
			Field: createdAt.Field,
		},
		IsPrimaryKey: createdAt.PrimaryKey,
		Value:        value,
	}
}
