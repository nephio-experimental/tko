package util

import (
	"github.com/tliron/go-ard"
)

func ToMapList(list []any) ([]ard.Map, bool) {
	list_ := make([]ard.Map, len(list))
	var ok bool
	for index, element := range list {
		if list_[index], ok = element.(ard.Map); !ok {
			return nil, false
		}
	}
	return list_, true
}
