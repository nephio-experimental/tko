package util

func StringSetAdd(set []string, entry string) ([]string, bool) {
	for _, entry_ := range set {
		if entry_ == entry {
			return set, false
		}
	}
	set = append(set, entry)
	return set, true
}

func StringSetRemove(set []string, entry string) ([]string, bool) {
	for index, entry_ := range set {
		if entry_ == entry {
			set = append(set[:index], set[index+1:]...)
			return set, true
		}
	}
	return set, false
}

func StringSetClone(set []string) []string {
	set_ := make([]string, len(set))
	copy(set_, set)
	return set_
}
