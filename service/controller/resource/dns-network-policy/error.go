package test

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidCRError = &microerror.Error{
	Kind: "invalidCRError",
}

// IsInvalidCR asserts invalidCRError.
func IsInvalidCR(err error) bool {
	return microerror.Cause(err) == invalidCRError
}
