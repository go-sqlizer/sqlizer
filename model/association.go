package model

import (
	"reflect"
)

type AssociationType uint8

const (
	HasManyAssociation AssociationType = iota
	HasOneAssociation
	BelongsToAssociation
	ManyToManyAssociation
)

type Association struct {
	Model      *Model
	Type       AssociationType
	Properties AssociationProperties
}

type AssociationProperties struct {
	ForeignKey string
	SourceKey  string
	TargetKey  string
	Through    *Model
}

func (m Model) AssociationFromModel(model Model) Association {
	// Get primaryKey
	associationType := reflect.ValueOf(m.Associations)
	for i := 0; i < associationType.NumField(); i++ {
		resultAssociation := associationType.Field(i).Interface().(Association)
		if *resultAssociation.Model == model {
			return resultAssociation
		}
	}

	panic("Invalid Model " + model.Name)
}
