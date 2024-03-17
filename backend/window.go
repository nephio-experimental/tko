package backend

var (
	DefaultMaxCount uint = 100
	MaxMaxCount     uint = 1000
)

//
// Window
//

type Window struct {
	Offset   uint
	MaxCount int // <0 for limitless
}

func (self Window) End() (uint, bool) {
	if self.MaxCount >= 0 {
		return self.Offset + uint(self.MaxCount), true
	} else {
		// Endless!
		return 0, false
	}
}

func ApplyWindow[E any](list []E, window Window) []E {
	length := uint(len(list))
	if window.Offset > length {
		return nil
	} else if end, limited := window.End(); !limited || (end > length) {
		return list[window.Offset:]
	} else {
		return list[window.Offset:end]
	}
}
