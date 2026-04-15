// Package variables provides access to Spin application variables.
package variables

import (
	"errors"

	variables "github.com/spinframework/spin-go-sdk/v3/imports/spin_variables_3_0_0_variables"
)

// Get returns an application variable value for the current component.
//
// The name must match one defined in the component manifest.
func Get(key string) (string, error) {
	result := variables.Get(key)
	if result.IsErr() {
		return "", errorVariantToError(result.Err())
	}

	return result.Ok(), nil
}

func errorVariantToError(err variables.Error) error {
	switch err.Tag() {
	case variables.ErrorInvalidName:
		return errors.New(err.InvalidName())
	case variables.ErrorProvider:
		return errors.New(err.Provider())
	case variables.ErrorUndefined:
		return errors.New(err.Undefined())
	case variables.ErrorOther:
		return errors.New(err.Other())
	default:
		return errors.New("no error provided by host implementation")
	}
}
