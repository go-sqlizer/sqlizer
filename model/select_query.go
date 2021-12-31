package model

import (
	"fmt"
	"github.com/Supersonido/sqlizer"
	"github.com/Supersonido/sqlizer/drivers"
	"github.com/Supersonido/sqlizer/queries"
	"reflect"
	"strings"
)

func SelectQuery(result reflect.Type, model Model, options queries.Options) drivers.Query {
	var columns []string
	var joins []string
	var extra []string
	tableAlias := model.Name

	// Render model fields
	for i := 0; i < result.NumField(); i++ {
		resultField := result.Field(i).Name

		if c := reflect.ValueOf(model.Columns).FieldByName(resultField); c.IsValid() {
			field := c.Interface().(Field)
			columns = append(columns, sqlizer.Conn.SelectColumn(tableAlias, field.Field, "", resultField))
		}
	}

	// Render associations fields
	for _, include := range options.Include {
		if a := reflect.ValueOf(model.Associations).FieldByName(include.As); a.IsValid() {
			association := a.Interface().(Association)
			var associationType *reflect.Type
			if associationTypeAux, ok := result.FieldByName(include.As); ok {
				associationType = typeResolver(associationTypeAux.Type)
			}

			newColumns, newJoins := generateAssociation(associationType, association, include, model, tableAlias)
			columns = append(columns, newColumns...)
			joins = append(joins, newJoins...)
		}
	}

	// Render Where
	if len(options.Where) > 0 {
		var fullWhere []string
		for _, where := range options.Where {
			fullWhere = append(fullWhere, string(where))
		}

		extra = append(extra, fmt.Sprintf("WHERE %s", strings.Join(fullWhere, " AND ")))
	}

	return drivers.Query{
		Statement: fmt.Sprintf(
			"SELECT %s \nFROM %s \n%s \n%s",
			strings.Join(columns, ",\n\t"),
			sqlizer.Conn.From(model.Schema, model.Table, tableAlias),
			strings.Join(joins, "\n"),
			strings.Join(extra, "\n"),
		),
		Values: []interface{}{},
	}
}

func generateAssociation(result *reflect.Type, association Association, options queries.Include, parent Model, parenAlias string) ([]string, []string) {
	var columns []string
	model := association.Model
	var tableAlias string
	// Get Table alias
	switch association.Type {
	case HasManyAssociation, ManyToManyAssociation:
		tableAlias = fmt.Sprintf("%s->%s", parenAlias, options.As)
	case BelongsToAssociation, HasOneAssociation:
		tableAlias = fmt.Sprintf("%s.%s", parenAlias, options.As)
		break
	}

	// Render model fields
	if result != nil {
		resultAux := *result

		for i := 0; i < resultAux.NumField(); i++ {
			resultField := resultAux.Field(i).Name

			if c := reflect.ValueOf(model.Columns).FieldByName(resultField); c.IsValid() {
				field := c.Interface().(Field)
				columns = append(columns, sqlizer.Conn.SelectColumn(tableAlias, field.Field, tableAlias, resultField))
			}
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
				if associationTypeAux, ok := resultAux.FieldByName(include.As); ok {
					associationType = typeResolver(associationTypeAux.Type)
				}
			}

			newColumns, newJoins := generateAssociation(associationType, childAssociation, include, *model, tableAlias)
			columns = append(columns, newColumns...)
			joins = append(joins, newJoins...)
		}
	}

	return columns, joins
}

func generateJoin(association Association, options queries.Include, tableAlias string, parent Model, parenAlias string) []string {
	var joinType drivers.JoinType
	var joins []string
	var parentField, childField string
	model := association.Model

	if !options.Required {
		joinType = drivers.LeftJoin
	} else {
		joinType = drivers.InnerJoin
	}

	switch association.Type {
	case BelongsToAssociation:
		parentField = parent.FieldFromName(association.Properties.ForeignKey).Field
		childField = association.Model.primaryKey.Field
	case HasManyAssociation:
		parentField = parent.primaryKey.Field
		childField = model.FieldFromName(association.Properties.ForeignKey).Field
	case ManyToManyAssociation:
		assoc := association.Properties.Through.AssociationFromModel(*association.Model)
		through := association.Properties.Through
		throughField := through.FieldFromName(association.Properties.ForeignKey).Field
		parentAliasAux := fmt.Sprintf("%s.%s", tableAlias, through.Name)
		joins = []string{
			sqlizer.Conn.Join(
				joinType,
				through.Schema, through.Table,
				parentAliasAux, throughField,
				parenAlias, parent.primaryKey.Field,
			),
		}
		parenAlias = parentAliasAux
		parentField = association.Properties.Through.FieldFromName(assoc.Properties.ForeignKey).Field
		childField = association.Model.primaryKey.Field
	}

	return append(joins, sqlizer.Conn.Join(joinType, model.Schema, model.Table, tableAlias, childField, parenAlias, parentField))
}

func typeResolver(p reflect.Type) *reflect.Type {
	switch p.Kind() {
	case reflect.Ptr, reflect.Array, reflect.Slice:
		return typeResolver(p.Elem())
	}

	return &p
}
