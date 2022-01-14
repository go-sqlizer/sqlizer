package common

import "github.com/go-sqlizer/sqlizer/queries"

func ContainsStr(list []string, findElem string) bool {
	for _, elem := range list {
		if elem == findElem {
			return true
		}
	}

	return false
}

func ContainsInclude(list []queries.Field, findElem string) bool {
	for _, elem := range list {
		if elem.As == findElem {
			return true
		}
	}

	return false
}
