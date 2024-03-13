package backend

import (
	"fmt"
)

//
// NotImplementedError
//

type NotImplementedError struct {
	message string
}

func NewNotImplementedError(feature string) *NotImplementedError {
	return &NotImplementedError{"not implemented: " + feature}
}

func IsNotImplementedError(err error) bool {
	_, ok := err.(*NotImplementedError)
	return ok
}

// (error interface)
func (self *NotImplementedError) Error() string {
	return self.message
}

//
// BadArgumentError
//

type BadArgumentError struct {
	message string
}

func NewBadArgumentError(message string) *BadArgumentError {
	if message == "" {
		return &BadArgumentError{"bad argument"}
	} else {
		return &BadArgumentError{"bad argument: " + message}
	}
}

func NewBadArgumentErrorf(format string, a ...any) *BadArgumentError {
	return NewBadArgumentError(fmt.Sprintf(format, a...))
}

func WrapBadArgumentError(err error) *BadArgumentError {
	return NewBadArgumentError(err.Error())
}

func IsBadArgumentError(err error) bool {
	_, ok := err.(*BadArgumentError)
	return ok
}

// (error interface)
func (self *BadArgumentError) Error() string {
	return self.message
}

//
// NotDoneError
//

type NotDoneError struct {
	message string
}

func NewNotDoneError(message string) *NotDoneError {
	if message == "" {
		return &NotDoneError{"not done"}
	} else {
		return &NotDoneError{"not done: " + message}
	}
}

func NewNotDoneErrorf(format string, a ...any) *NotDoneError {
	return NewNotDoneError(fmt.Sprintf(format, a...))
}

func IsNotDoneError(err error) bool {
	_, ok := err.(*NotDoneError)
	return ok
}

// (error interface)
func (self *NotDoneError) Error() string {
	return self.message
}

//
// NotFoundError
//

type NotFoundError struct {
	message string
}

func NewNotFoundError(message string) *NotFoundError {
	if message == "" {
		return &NotFoundError{"not found"}
	} else {
		return &NotFoundError{"not found: " + message}
	}
}

func NewNotFoundErrorf(format string, a ...any) *NotFoundError {
	return NewNotFoundError(fmt.Sprintf(format, a...))
}

func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// (error interface)
func (self *NotFoundError) Error() string {
	return self.message
}

//
// BusyError
//

type BusyError struct {
	message string
}

func NewBusyError(message string) *BusyError {
	if message == "" {
		return &BusyError{"busy"}
	} else {
		return &BusyError{"busy: " + message}
	}
}

func NewBusyErrorf(format string, a ...any) *BusyError {
	return NewBusyError(fmt.Sprintf(format, a...))
}

func IsBusyError(err error) bool {
	_, ok := err.(*BusyError)
	return ok
}

// (error interface)
func (self *BusyError) Error() string {
	return self.message
}

//
// TimeoutError
//

type TimeoutError struct {
	message string
}

func NewTimeoutError(message string) *TimeoutError {
	if message == "" {
		return &TimeoutError{"timeout"}
	} else {
		return &TimeoutError{"timeout: " + message}
	}
}

func NewTimeoutErrorf(format string, a ...any) *TimeoutError {
	return NewTimeoutError(fmt.Sprintf(format, a...))
}

func IsTimeoutError(err error) bool {
	_, ok := err.(*TimeoutError)
	return ok
}

// (error interface)
func (self *TimeoutError) Error() string {
	return self.message
}
