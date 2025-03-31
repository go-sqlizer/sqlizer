package model

import (
	"fmt"
	"github.com/go-sqlizer/sqlizer/common"
	"github.com/go-sqlizer/sqlizer/queries"
	"reflect"
)

func SelectBuilder(result reflect.Type, model *Model, options queries.QueryOptions) queries.BasicQuery {
	var columns []queries.Column
	var joins []queries.Join
	tableAlias := model.Name

	// Render model fields
	for i := 0; i < result.NumField(); i++ {
		resultField := result.Field(i)

		// Exclude Fields
		if common.ContainsStr(options.Fields.Excludes, resultField.Name) || common.ContainsInclude(options.Fields.Includes, resultField.Name) {
			continue
		}

		if c := reflect.ValueOf(model.Columns).FieldByName(resultField.Name); c.IsValid() {
			field := c.Interface().(Field)

			resultType := &resultField.Type
			columnType := reflect.Type(field.Type)

			columns = append(columns, queries.Column{
				Alias:        resultField.Name,
				Type:         resultType,
				ColumnType:   &columnType,
				IsPrimaryKey: field.PrimaryKey,
				Get:          field.Get,
				Set:          field.Set,
				Source: &queries.ColumnSource{
					Alias: tableAlias,
					Field: field.Field,
				},
			})
		}
	}

	// Render Extra fields
	for _, field := range options.Fields.Includes {
		var source *queries.ColumnSource
		var fColumnType reflect.Type
		var isPk bool

		if c := reflect.ValueOf(model.Columns).FieldByName(field.As); c.IsValid() {
			fColumn := c.Interface().(Field)
			fColumnType = reflect.Type(fColumn.Type)
			isPk = fColumn.PrimaryKey
			source = &queries.ColumnSource{
				Alias: tableAlias,
				Field: fColumn.Field,
			}
		} else if c, ok := result.FieldByName(field.As); ok {
			fColumnType = c.Type
		} else if field.Fn != nil {
			fColumnType = *field.Fn.Type
		} else {
			panic(fmt.Sprintf("Missing type for field %s.%s", tableAlias, field.As))
		}

		columns = append(columns, queries.Column{
			Alias:        field.As,
			Function:     field.Fn,
			Type:         &fColumnType,
			ColumnType:   &fColumnType,
			IsPrimaryKey: isPk,
			Source:       source,
		})
	}

	// Render associations fields
	for _, include := range options.Include {
		if a := reflect.ValueOf(model.Associations).FieldByName(include.As); a.IsValid() {
			association := a.Interface().(Association)
			var associationType *reflect.Type
			associationTypeAux, ok := result.FieldByName(include.As)

			if ok && !common.ContainsStr(options.Fields.Excludes, include.As) && !common.ContainsInclude(options.Fields.Includes, include.As) {
				associationType = common.TypeResolver(associationTypeAux.Type)
			}

			newColumns, newJoins := generateAssociation(associationType, association, include, model, tableAlias)
			joins = append(joins, newJoins...)
			if associationType != nil {
				columns = append(columns, queries.Column{
					Alias:  include.As,
					Nested: &newColumns,
				})
			}
		}
	}

	return queries.BasicQuery{
		QueryOptions: options,
		Columns:      columns,
		Joins:        joins,
		From: queries.TableSource{
			Schema: model.Schema,
			Table:  model.Table,
			Alias:  tableAlias,
		},
	}
}

func generateAssociation(result *reflect.Type, association Association, options queries.Include, parent *Model, parenAlias string) ([]queries.Column, []queries.Join) {
	var columns []queries.Column
	model := association.Model
	tableAlias := fmt.Sprintf("%s.%s", parenAlias, options.As)

	if result != nil {
		resultAux := *result

		// Render model fields
		for i := 0; i < resultAux.NumField(); i++ {
			resultField := resultAux.Field(i)

			// Exclude Fields
			if common.ContainsStr(options.Fields.Excludes, resultField.Name) || common.ContainsInclude(options.Fields.Includes, resultField.Name) {
				continue
			}

			if c := reflect.ValueOf(model.Columns).FieldByName(resultField.Name); c.IsValid() {
				field := c.Interface().(Field)
				columnType := reflect.Type(field.Type)

				columns = append(columns, queries.Column{
					Alias:        resultField.Name,
					Type:         &resultField.Type,
					ColumnType:   &columnType,
					IsPrimaryKey: field.PrimaryKey,
					Get:          field.Get,
					Set:          field.Set,
					Source: &queries.ColumnSource{
						Alias: tableAlias,
						Field: field.Field,
					},
				})
			}
		}

		// Render Extra fields
		for _, field := range options.Fields.Includes {
			var source *queries.ColumnSource
			var fColumnType reflect.Type
			var isPk bool

			if c := reflect.ValueOf(model.Columns).FieldByName(field.As); c.IsValid() {
				fColumn := c.Interface().(Field)
				fColumnType = reflect.Type(fColumn.Type)
				isPk = fColumn.PrimaryKey
				source = &queries.ColumnSource{
					Alias: tableAlias,
					Field: fColumn.Field,
				}
			} else if c, ok := resultAux.FieldByName(field.As); ok {
				fColumnType = c.Type
			} else if field.Fn != nil {
				fColumnType = *field.Fn.Type
			} else {
				panic(fmt.Sprintf("Missing type for field %s.%s", tableAlias, field.As))
			}

			columns = append(columns, queries.Column{
				Alias:        field.As,
				Function:     field.Fn,
				Type:         &fColumnType,
				ColumnType:   &fColumnType,
				IsPrimaryKey: isPk,
				Source:       source,
			})
		}
	}

	// Render associations fields
	joins := generateJoin(association, options, tableAlias, parent, parenAlias)
	for _, include := range options.Include {
		if a := reflect.ValueOf(model.Associations).FieldByName(include.As); a.IsValid() {
			childAssociation := a.Interface().(Association)
			var associationType *reflect.Type
			if result != nil {
				resultAux := *result
				associationTypeAux, ok := resultAux.FieldByName(include.As)

				if ok && !common.ContainsStr(options.Fields.Excludes, include.As) && !common.ContainsInclude(options.Fields.Includes, include.As) {
					associationType = common.TypeResolver(associationTypeAux.Type)
				}
			}

			newColumns, newJoins := generateAssociation(associationType, childAssociation, include, model, tableAlias)
			joins = append(joins, newJoins...)
			if associationType != nil {
				columns = append(columns, queries.Column{
					Alias:  include.As,
					Nested: &newColumns,
				})
			}
		}
	}

	return columns, joins
}

func generateJoin(association Association, options queries.Include, tableAlias string, parent *Model, parenAlias string) []queries.Join {
	model := association.Model

	primaryKey := queries.ColumnValue{Alias: parenAlias, Field: parent.primaryKey.Field}
	if association.Properties.SourceKey != "" {
		primaryKey = queries.ColumnValue{Alias: parenAlias, Field: parent.FieldFromName(association.Properties.SourceKey).Field}
	}

	switch association.Type {
	case BelongsToAssociation:
		return []queries.Join{
			{
				Type: options.JoinType,
				From: parenAlias,
				To: queries.TableSource{
					Schema: model.Schema,
					Table:  model.Table,
					Alias:  tableAlias,
				},
				Where: append(
					[]queries.Where{
						queries.Eq(
							queries.ColumnValue{Alias: parenAlias, Field: parent.FieldFromName(association.Properties.ForeignKey).Field},
							queries.ColumnValue{Alias: tableAlias, Field: association.Model.primaryKey.Field},
						),
					},
					options.Where...,
				),
			},
		}
	case HasManyAssociation:
		return []queries.Join{
			{
				Type: options.JoinType,
				From: parenAlias,
				To: queries.TableSource{
					Schema: model.Schema,
					Table:  model.Table,
					Alias:  tableAlias,
				},
				Where: append(
					[]queries.Where{
						queries.Eq(
							primaryKey,
							queries.ColumnValue{Alias: tableAlias, Field: model.FieldFromName(association.Properties.ForeignKey).Field},
						),
					},
					options.Where...,
				),
			},
		}
	case HasOneAssociation:
		return []queries.Join{
			{
				Type: options.JoinType,
				From: parenAlias,
				To: queries.TableSource{
					Schema: model.Schema,
					Table:  model.Table,
					Alias:  tableAlias,
				},
				Where: append(
					[]queries.Where{
						queries.Eq(
							primaryKey,
							queries.ColumnValue{Alias: tableAlias, Field: model.FieldFromName(association.Properties.ForeignKey).Field},
						),
					},
					options.Where...,
				),
			},
		}
	case ManyToManyAssociation:
		assoc := association.Properties.Through.AssociationFromModel(*association.Model)
		through := association.Properties.Through
		parentAliasAux := fmt.Sprintf("%s.%s", tableAlias, through.Name)
		return []queries.Join{
			{
				Type: options.JoinType,
				From: parenAlias,
				To: queries.TableSource{
					Schema: through.Schema,
					Table:  through.Table,
					Alias:  parentAliasAux,
				},
				Where: []queries.Where{
					queries.Eq(
						queries.ColumnValue{Alias: parenAlias, Field: parent.primaryKey.Field},
						queries.ColumnValue{Alias: parentAliasAux, Field: through.FieldFromName(association.Properties.ForeignKey).Field},
					),
				},
			},
			{
				Type: options.JoinType,
				From: parentAliasAux,
				To: queries.TableSource{
					Schema: model.Schema,
					Table:  model.Table,
					Alias:  tableAlias,
				},
				Where: append(
					[]queries.Where{
						queries.Eq(
							queries.ColumnValue{Alias: parentAliasAux, Field: association.Properties.Through.FieldFromName(assoc.Properties.ForeignKey).Field},
							queries.ColumnValue{Alias: tableAlias, Field: association.Model.primaryKey.Field},
						),
					},
					options.Where...,
				),
			},
		}
	}

	return []queries.Join{}
}
