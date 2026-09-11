package validation_test

import (
	"reflect"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/validation"
)

type person struct{ age *int }

func (p *person) GetAge() *int { return p.age }

func (p *person) Constraints() []validation.Constraint {
	return []validation.Constraint{
		validation.Max("age", 90, "too old"),
	}
}

func age(n int) *int { return &n }

// TestJSR380Validation covers the constraint evaluation the controller runs
// per request, including JSR-380's rule that a null value satisfies @Max and
// @Min but not @NotNull.
func TestJSR380Validation(t *testing.T) {
	v := validation.BuildDefaultValidatorFactory()
	cases := []struct {
		name  string
		value *int
		want  []string
	}{
		{"in range", age(30), nil},
		{"at the maximum", age(90), nil},
		{"over the maximum", age(91), []string{"too old"}},
		{"zero", age(0), nil},
		{"absent", nil, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := v.Validate(&person{age: tc.value})
			if len(got) != len(tc.want) {
				t.Fatalf("got %d violations %v, want %d %v", len(got), messages(got), len(tc.want), tc.want)
			}
			for i := range tc.want {
				if got[i].GetMessage() != tc.want[i] {
					t.Errorf("violation %d = %q, want %q", i, got[i].GetMessage(), tc.want[i])
				}
				if got[i].GetPropertyPath() != "age" {
					t.Errorf("property path = %q", got[i].GetPropertyPath())
				}
			}
		})
	}

	if got := v.Validate("not constrained"); got != nil {
		t.Errorf("an unconstrained value produced %v", got)
	}
}

func messages(vs []validation.ConstraintViolation) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.GetMessage()
	}
	return out
}

// rejecting is a Spring Validator shaped like the application's own.
type rejecting struct{}

func (rejecting) Supports(t reflect.Type) bool { return t == reflect.TypeOf(&person{}) }
func (rejecting) Validate(target any, errors *validation.Errors) {
	if target.(*person).age == nil {
		errors.Reject("100", "invalid")
	}
}

// TestSpringValidatorAndErrors covers the Errors surface a Spring Validator
// writes into.
func TestSpringValidatorAndErrors(t *testing.T) {
	var v validation.Validator = rejecting{}
	if !v.Supports(reflect.TypeOf(&person{})) || v.Supports(reflect.TypeOf(person{})) {
		t.Error("Supports does not distinguish the exact type")
	}

	errs := validation.NewErrors("person")
	if errs.HasErrors() {
		t.Error("a fresh Errors reports errors")
	}
	if errs.GetObjectName() != "person" {
		t.Errorf("object name = %q", errs.GetObjectName())
	}
	v.Validate(&person{}, errs)
	if !errs.HasErrors() {
		t.Fatal("the rejection was not recorded")
	}
	all := errs.GetAllErrors()
	if len(all) != 1 || all[0].Code != "100" || all[0].DefaultMessage != "invalid" {
		t.Errorf("errors = %+v", all)
	}

	v.Validate(&person{age: age(1)}, validation.NewErrors("person"))
}

// TestBindingResult covers the Errors implementation a controller receives.
func TestBindingResult(t *testing.T) {
	target := &person{}
	br := validation.NewBindingResult(target, "simpleModel")
	if br.GetTarget() != any(target) {
		t.Error("GetTarget does not report the bound object")
	}
	if br.HasErrors() {
		t.Error("a fresh BindingResult reports errors")
	}
	br.Reject("100", "invalid")
	if !br.HasErrors() {
		t.Error("Reject was not recorded")
	}
}
