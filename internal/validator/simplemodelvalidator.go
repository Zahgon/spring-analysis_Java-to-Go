// Package validator holds the application's own Spring Validator.
package validator

import (
	"reflect"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/validation"
	"github.com/seaswalker/spring-analysis/internal/model"
)

// The rejection the validator records.
const (
	// AgeErrorCode is the error code the original rejects with.
	AgeErrorCode = "100"
	// AgeErrorMessage is the default message the original rejects with.
	AgeErrorMessage = "年龄不合法"
	// AgeMin and AgeMax bound the accepted age, exclusive of the values just
	// outside them.
	AgeMin = 1
	AgeMax = 200
)

// SimpleModelValidator rejects a SimpleModel whose age is absent or outside
// the accepted range.
//
// Nothing registers it: the @InitBinder lines that would are commented out in
// the original, so it never runs in the web flow. It is ported with its bounds
// and message intact.
type SimpleModelValidator struct{}

// NewSimpleModelValidator returns the validator.
func NewSimpleModelValidator() *SimpleModelValidator { return &SimpleModelValidator{} }

// Supports accepts exactly SimpleModel, as the original's `clazz ==
// SimpleModel.class` does — a subclass would not be accepted either.
func (v *SimpleModelValidator) Supports(t reflect.Type) bool {
	return t == reflect.TypeOf(&model.SimpleModel{}) || t == reflect.TypeOf(model.SimpleModel{})
}

// Validate rejects an age that is absent, below AgeMin, or above AgeMax.
func (v *SimpleModelValidator) Validate(target any, errors *validation.Errors) {
	simpleModel, ok := target.(*model.SimpleModel)
	if !ok {
		return
	}
	age := simpleModel.GetAge()
	if age == nil || *age < AgeMin || *age > AgeMax {
		errors.Reject(AgeErrorCode, AgeErrorMessage)
	}
}

var _ validation.Validator = (*SimpleModelValidator)(nil)
