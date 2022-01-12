package common

func ContainsStr(list []string, findElem string) bool {
	for _, elem := range list {
		if elem == findElem {
			return true
		}
	}

	return false
}
