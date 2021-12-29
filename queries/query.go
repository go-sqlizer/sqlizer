package queries

type Options struct {
	Logging func(...interface{})
	Include []Include
	Where   []WhereOption
	Limit   *int
	Offset  *int
}

type IncludeOperations uint8

type Include struct {
	As       string
	Include  []Include
	Where    []WhereOption
	Required bool
}

type PaginateOptions struct {
	Options
	Page    int
	PerPage int
}
