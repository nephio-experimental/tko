package backend

var (
	DefaultMaxCount uint = 100
	MaxMaxCount     uint = 1_000
)

//
// Window
//

type Window struct {
	Offset   uint
	MaxCount int // <0 for maximum number of results
}

func (self Window) Limit() uint {
	if self.MaxCount >= 0 {
		return uint(self.MaxCount)
	} else {
		return MaxMaxCount
	}
}

func (self Window) End() uint {
	return self.Offset + self.Limit()
}

func ApplyWindow[E any](list []E, window Window) []E {
	length := uint(len(list))
	if window.Offset > length {
		return nil
	} else if end := window.End(); end > length {
		return list[window.Offset:]
	} else {
		return list[window.Offset:end]
	}
}
