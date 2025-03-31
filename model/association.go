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

func (model *Model) AssociationFromModel(assocModel Model) Association {
	associationType := reflect.ValueOf(model.Associations)
	for i := 0; i < associationType.NumField(); i++ {
		resultAssociation := associationType.Field(i).Interface().(Association)
		if resultAssociation.Model.Name == assocModel.Name {
			return resultAssociation
		}
	}

	panic("Invalid Model " + assocModel.Name)
}
