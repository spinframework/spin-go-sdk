// Code generated by wit-bindgen-go. DO NOT EDIT.

// Package variables represents the imported interface "fermyon:spin/variables@2.0.0".
package variables

import (
	"go.bytecodealliance.org/cm"
)

// Error represents the variant "fermyon:spin/variables@2.0.0#error".
//
// The set of errors which may be raised by functions in this interface.
//
//	variant error {
//		invalid-name(string),
//		undefined(string),
//		provider(string),
//		other(string),
//	}
type Error cm.Variant[uint8, string, string]

// ErrorInvalidName returns a [Error] of case "invalid-name".
//
// The provided variable name is invalid.
func ErrorInvalidName(data string) Error {
	return cm.New[Error](0, data)
}

// InvalidName returns a non-nil *[string] if [Error] represents the variant case "invalid-name".
func (self *Error) InvalidName() *string {
	return cm.Case[string](self, 0)
}

// ErrorUndefined returns a [Error] of case "undefined".
//
// The provided variable is undefined.
func ErrorUndefined(data string) Error {
	return cm.New[Error](1, data)
}

// Undefined returns a non-nil *[string] if [Error] represents the variant case "undefined".
func (self *Error) Undefined() *string {
	return cm.Case[string](self, 1)
}

// ErrorProvider returns a [Error] of case "provider".
//
// A variables provider specific error has occurred.
func ErrorProvider(data string) Error {
	return cm.New[Error](2, data)
}

// Provider returns a non-nil *[string] if [Error] represents the variant case "provider".
func (self *Error) Provider() *string {
	return cm.Case[string](self, 2)
}

// ErrorOther returns a [Error] of case "other".
//
// Some implementation-specific error has occurred.
func ErrorOther(data string) Error {
	return cm.New[Error](3, data)
}

// Other returns a non-nil *[string] if [Error] represents the variant case "other".
func (self *Error) Other() *string {
	return cm.Case[string](self, 3)
}

var stringsError = [4]string{
	"invalid-name",
	"undefined",
	"provider",
	"other",
}

// String implements [fmt.Stringer], returning the variant case name of v.
func (v Error) String() string {
	return stringsError[v.Tag()]
}

// Get represents the imported function "get".
//
// Get an application variable value for the current component.
//
// The name must match one defined in in the component manifest.
//
//	get: func(name: string) -> result<string, error>
//
//go:nosplit
func Get(name string) (result cm.Result[ErrorShape, string, Error]) {
	name0, name1 := cm.LowerString(name)
	wasmimport_Get((*uint8)(name0), (uint32)(name1), &result)
	return
}
