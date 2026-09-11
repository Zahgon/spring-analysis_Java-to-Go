package validator_test

import (
	"reflect"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/validation"
	"github.com/seaswalker/spring-analysis/internal/model"
	"github.com/seaswalker/spring-analysis/internal/validator"
)

func aged(n int) *model.SimpleModel {
	m := model.NewSimpleModel()
	m.SetAge(&n)
	return m
}

// TestSupportsExactlyTheModelType covers `clazz == SimpleModel.class`.
func TestSupportsExactlyTheModelType(t *testing.T) {
	v := validator.NewSimpleModelValidator()
	if !v.Supports(reflect.TypeOf(&model.SimpleModel{})) {
		t.Error("the model type is not supported")
	}
	if v.Supports(reflect.TypeOf("")) {
		t.Error("an unrelated type is supported")
	}
}

// TestAgeBounds covers the range the validator accepts and the rejection it
// records outside it.
func TestAgeBounds(t *testing.T) {
	v := validator.NewSimpleModelValidator()
	cases := []struct {
		name       string
		target     *model.SimpleModel
		wantReject bool
	}{
		{"absent", model.NewSimpleModel(), true},
		{"below the minimum", aged(0), true},
		{"at the minimum", aged(validator.AgeMin), false},
		{"in range", aged(30), false},
		{"at the maximum", aged(validator.AgeMax), false},
		{"above the maximum", aged(validator.AgeMax + 1), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := validation.NewErrors("simpleModel")
			v.Validate(tc.target, errs)
			if errs.HasErrors() != tc.wantReject {
				t.Fatalf("rejected = %v, want %v", errs.HasErrors(), tc.wantReject)
			}
			if !tc.wantReject {
				return
			}
			all := errs.GetAllErrors()
			if all[0].Code != validator.AgeErrorCode || all[0].DefaultMessage != validator.AgeErrorMessage {
				t.Errorf("rejection = %+v, want code %q and message %q",
					all[0], validator.AgeErrorCode, validator.AgeErrorMessage)
			}
		})
	}
}

// TestValidateIgnoresOtherTypes covers the guard on the cast.
func TestValidateIgnoresOtherTypes(t *testing.T) {
	errs := validation.NewErrors("other")
	validator.NewSimpleModelValidator().Validate("not a model", errs)
	if errs.HasErrors() {
		t.Error("an unrelated target was rejected")
	}
}
