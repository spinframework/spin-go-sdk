package variables

import (
	"fmt"

	variables "github.com/spinframework/spin-go-sdk/v3/internal/fermyon_spin_2_0_0_variables"
)

// Get an application variable value for the current component.
//
// The name must match one defined in in the component manifest.
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
		return fmt.Errorf("%v", err.InvalidName())
	case variables.ErrorProvider:
		return fmt.Errorf("%v", err.Provider())
	case variables.ErrorUndefined:
		return fmt.Errorf("%v", err.Undefined())
	case variables.ErrorOther:
		return fmt.Errorf("%v", err.Other())
	default:
		return fmt.Errorf("no error provided by host implementation")
	}
}
