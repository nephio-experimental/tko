package backend

import (
	"io"
)

//
// Results
//

type Results[E any] interface {
	Next() (E, error) // error can be io.EOF
}

//
// ResultsSlice
//

type ResultsSlice[E any] struct {
	entities []E
	length   int
	index    int
}

func NewResultsSlice[E any](entities []E) *ResultsSlice[E] {
	return &ResultsSlice[E]{
		entities: entities,
		length:   len(entities),
	}
}

// ([Results] interface)
func (self *ResultsSlice[E]) Next() (E, error) {
	if self.index < self.length {
		entity := self.entities[self.index]
		self.index++
		return entity, nil
	} else {
		return *new(E), io.EOF
	}
}

//
// ResultsStream
//

const FEED_STREAM_SIZE = 100

type ResultsStream[E any] struct {
	entities chan E
	errors   chan error
}

func NewResultsStream[E any]() *ResultsStream[E] {
	return &ResultsStream[E]{
		entities: make(chan E, FEED_STREAM_SIZE),
		errors:   make(chan error),
	}
}

// ([Results] interface)
func (self *ResultsStream[E]) Next() (E, error) {
	for {
		select {
		case entity, ok := <-self.entities:
			if ok {
				return entity, nil
			} else {
				return *new(E), io.EOF
			}

		case err := <-self.errors:
			return *new(E), err
		}
	}
}

func (self *ResultsStream[E]) Close(err error) {
	if err == nil {
		close(self.entities)
	} else {
		self.errors <- err
	}
}

func (self *ResultsStream[E]) Send(info E) {
	self.entities <- info
}
