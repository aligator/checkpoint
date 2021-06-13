// Package checkpoint provides a way to decorate errors by some additional caller information
// which results in something similar to a stacktrace.
// Each error added to a Checkpoint can be checked by errors.Is and retrieved by errors.As.
package checkpoint

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
)

// IgnoreEOF returns the io.EOF and io.ErrUnexpectedEOF directly instead of wrapping it.
// This may be needed to be compatible to several io functions from the standard lib and from other libs.
// These often check for io.EOF by equality and not by errors.Is because of historical reasons.
// See https://github.com/golang/go/issues/39155
func IgnoreEOF() Option {
	return func(err error) error {
		if err == io.EOF {
			return io.EOF
		}
		if err == io.ErrUnexpectedEOF {
			return io.ErrUnexpectedEOF
		}

		return nil
	}
}

// Option defines a special error handling function.
// If it returns nil the normal error handling is done.
// Else the returned error is returned instead of the Checkpoint.
//
// This can be used to define special error handling for special errors
// such as io.EOF.
type Option = func(err error) error

// From just wraps an error by a new Checkpoint which adds some caller information to the error.
// It returns nil, if err == nil.
// You may use Options to change the resulting error for some specific input-errors.
// (Such as IgnoreEOF for special EOF handling)
func From(err error, options ...Option) error {
	for _, o := range options {
		if newErr := o(err); newErr != nil {
			return newErr
		}
	}

	if err == nil {
		return nil
	}

	// Get the caller information.
	_, file, line, ok := runtime.Caller(1)

	return Checkpoint{
		err:  err,
		prev: nil,

		callerOk: ok,
		file:     filepath.Base(file),
		line:     line,
	}
}

// Wrap adds a Checkpoint with some caller information from an error and accepts
// also another error which can further describe the Checkpoint.
//
// You may use Options to change the resulting error for some specific input-errors.
// (Such as IgnoreEOF for special EOF handling)
//
// Returns nil if prev == nil.
// If err is nil, it still creates a Checkpoint.
// This allows for example to predefine some errors and use them later:
//  var(
//  		ErrSomethingSpecialWentWrong = errors.New("a very bad error")
//  )
//  func someFunction() error {
//  	err := somethingOtherThatThrowsErrors()
//  	return checkpoint.Wrap(err, ErrSomethingSpecialWentWrong)
//  }
//
//  err := someFunction()
// If used that way, you can still check with errors.Is() and errors.As() for the ErrSomethingSpecialWentWrong
//  if errors.Is(err, ErrSomethingSpecialWentWrong) {
//  	fmt.Println("The special error was thrown")
//  } else {
//  	fmt.Println(err)
//  }
// but also for the error returned by somethingOtherThatThrowsErrors() (if you know what error it is).
// If the error in this example is nil, no Checkpoint gets created.
func Wrap(prev, err error, options ...Option) error {
	for _, o := range options {
		if newErr := o(err); newErr != nil {
			return newErr
		}
	}

	if prev == nil {
		return nil
	}

	// Get the caller information.
	_, file, line, ok := runtime.Caller(1)

	return Checkpoint{
		err:  err,
		prev: prev,

		callerOk: ok,
		file:     filepath.Base(file),
		line:     line,
	}
}

type Checkpoint struct {
	err  error
	prev error

	callerOk bool
	file     string
	line     int
}

func (e Checkpoint) Error() string {
	prevErrString := ""
	if e.prev != nil {
		// Use different formatting for the prev error if it was not also a Checkpoint.
		prevErrString = e.prev.Error()
		_, ok := e.prev.(*Checkpoint)
		if !ok {
			prevErrString = "File: unknown\n\t" + strings.ReplaceAll(prevErrString, "\n", "\n\t")
		}
	}

	// Format different based on existing caller information.
	if e.callerOk {
		return fmt.Sprintf("File: %s:%d\n\t%v\n%v", e.file, e.line, e.err, prevErrString)
	}
	return fmt.Sprintf("File: unknown\n\t%v\n%v", e.err, prevErrString)
}

func (e Checkpoint) Unwrap() error {
	return e.prev
}

func (e Checkpoint) Is(target error) bool {
	return errors.Is(e.err, target)
}

func (e Checkpoint) As(target interface{}) bool {
	return errors.As(e.err, target)
}

func (e Checkpoint) File() string {
	return e.file
}

func (e Checkpoint) Line() int {
	return e.line
}
