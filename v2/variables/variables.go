package variables

import (
	"fmt"

	"github.com/spinframework/spin-go-sdk/v2/internal/fermyon/spin/v2.0.0/variables"
)

// Get an application variable value for the current component.
//
// The name must match one defined in in the component manifest.
func Get(key string) (string, error) {
	result := variables.Get(key)
	if result.IsErr() {
		return "", errorVariantToError(*result.Err())
	}

	return *result.OK(), nil
}

func errorVariantToError(err variables.Error) error {
	switch {
	case err.InvalidName() != nil:
		return fmt.Errorf(*err.InvalidName())
	case err.Provider() != nil:
		return fmt.Errorf(*err.Provider())
	case err.Undefined() != nil:
		return fmt.Errorf(*err.Undefined())
	default:
		if err.Other() != nil {
			return fmt.Errorf(*err.Other())
		}

		return fmt.Errorf("no error provided by host implementation")
	}
}
