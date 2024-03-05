package util

func CloneStringList(list []string) []string {
	set_ := make([]string, len(list))
	copy(set_, list)
	return set_
}

func AddToStringSet(set []string, entry string) ([]string, bool) {
	for _, entry_ := range set {
		if entry_ == entry {
			return set, false
		}
	}
	set = append(set, entry)
	return set, true
}

func RemoveFromStringSet(set []string, entry string) ([]string, bool) {
	for index, entry_ := range set {
		if entry_ == entry {
			set = append(set[:index], set[index+1:]...)
			return set, true
		}
	}
	return set, false
}

func CloneStringMap(map_ map[string]string) map[string]string {
	map__ := make(map[string]string)
	if map_ != nil {
		for key, value := range map_ {
			map__[key] = value
		}
	}
	return map__
}

func StringMapEquals(map1 map[string]string, map2 map[string]string) bool {
	if map1 == nil {
		return (map2 == nil) || (len(map2) == 0)
	}

	if map2 == nil {
		return (map1 == nil) || (len(map1) == 0)
	}

	if len(map1) != len(map2) {
		return false
	}

	for key1, value1 := range map1 {
		if value2, ok := map2[key1]; ok {
			if value1 != value2 {
				return false
			}
		} else {
			return false
		}
	}

	return true
}
