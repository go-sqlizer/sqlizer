package model

import (
	"reflect"
)

type Model struct {
	Name         string
	Schema       string
	Table        string
	Columns      interface{}
	Associations interface{}
	primaryKey   *Field
}

type Field struct {
	Field        string
	Get          *func(model interface{}) interface{}
	Set          *func(model interface{}, value interface{}) interface{}
	AllowNull    bool
	PrimaryKey   bool
	DefaultValue interface{}
	VirtualField bool
}

func (m *Model) Init() {
	// Get primaryKey
	columnsType := reflect.ValueOf(m.Columns)
	for i := 0; i < columnsType.NumField(); i++ {
		resultField := columnsType.Field(i).Interface().(Field)
		if resultField.PrimaryKey {
			m.primaryKey = &resultField
		}
	}
}

func (m Model) FieldFromName(name string) Field {
	columnsType := reflect.ValueOf(m.Columns)
	return columnsType.FieldByName(name).Interface().(Field)
}
