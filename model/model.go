package model

import (
	"github.com/go-sqlizer/sqlizer/drivers"
	"github.com/go-sqlizer/sqlizer/queries"
	"github.com/go-sqlizer/sqlizer/types"
	"reflect"
)

type Model struct {
	Name         string
	Schema       string
	Table        string
	Columns      interface{}
	Associations interface{}
	primaryKey   *Field
	driver       drivers.Driver
}

type Field struct {
	Field        string
	Type         types.FieldType
	Get          func(model interface{}) interface{}
	Set          func(model interface{}, value interface{}) interface{}
	AllowNull    bool
	PrimaryKey   bool
	DefaultValue interface{}
	VirtualField bool
}

func (model *Model) Init(driver drivers.Driver) *Model {
	// Find PrimaryKey
	columnsType := reflect.ValueOf(model.Columns)
	for i := 0; i < columnsType.NumField(); i++ {
		resultField := columnsType.Field(i).Interface().(Field)
		if resultField.PrimaryKey {
			model.primaryKey = &resultField
			break
		}
	}

	// Save driver
	model.driver = driver

	return model
}

func (model Model) FieldFromName(name string) Field {
	columnsType := reflect.ValueOf(model.Columns)
	return columnsType.FieldByName(name).Interface().(Field)
}

func (model Model) GetTableName() string {
	return model.driver.SerializeTableSource(queries.TableSource{
		Schema: model.Schema,
		Table:  model.Table,
	})
}
