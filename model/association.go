package model

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
