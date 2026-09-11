// Package validation reproduces the two validation surfaces the application
// uses: Spring's own org.springframework.validation.Validator/Errors, and the
// JSR-380 subset hibernate-validator supplies for the constraints declared on
// the model.
package validation

import (
	"reflect"
	"sort"
	"strings"
)

// ObjectError is one rejection recorded against an object, as
// org.springframework.validation.ObjectError.
type ObjectError struct {
	ObjectName     string
	Code           string
	DefaultMessage string
}

// Errors collects rejections, mirroring
// org.springframework.validation.Errors.
type Errors struct {
	objectName string
	errors     []ObjectError
}

// NewErrors returns an empty Errors for the named object.
func NewErrors(objectName string) *Errors {
	return &Errors{objectName: objectName}
}

// GetObjectName returns the name of the object being validated.
func (e *Errors) GetObjectName() string { return e.objectName }

// Reject records a global rejection with an error code and a default message.
func (e *Errors) Reject(code, defaultMessage string) {
	e.errors = append(e.errors, ObjectError{ObjectName: e.objectName, Code: code, DefaultMessage: defaultMessage})
}

// HasErrors reports whether anything was rejected.
func (e *Errors) HasErrors() bool { return len(e.errors) > 0 }

// GetAllErrors returns every rejection, in the order they were recorded.
func (e *Errors) GetAllErrors() []ObjectError {
	out := make([]ObjectError, len(e.errors))
	copy(out, e.errors)
	return out
}

// Validator mirrors org.springframework.validation.Validator.
type Validator interface {
	Supports(t reflect.Type) bool
	Validate(target any, errors *Errors)
}

// BindingResult is the Errors implementation a controller receives beside its
// bound model, as org.springframework.validation.BindingResult.
type BindingResult struct {
	*Errors
	target any
}

// NewBindingResult returns a BindingResult for a bound target.
func NewBindingResult(target any, objectName string) *BindingResult {
	return &BindingResult{Errors: NewErrors(objectName), target: target}
}

// GetTarget returns the bound object.
func (b *BindingResult) GetTarget() any { return b.target }

// ---------------------------------------------------------------------------
// JSR-380
// ---------------------------------------------------------------------------

// ConstraintViolation mirrors javax.validation.ConstraintViolation, reduced to
// what the controller reads from it.
type ConstraintViolation struct {
	PropertyPath string
	Message      string
	InvalidValue any
}

// GetMessage returns the interpolated constraint message.
func (v ConstraintViolation) GetMessage() string { return v.Message }

// GetPropertyPath returns the path of the property that failed.
func (v ConstraintViolation) GetPropertyPath() string { return v.PropertyPath }

// Constraint is one declared constraint on one property.
type Constraint struct {
	// Property is the property path the violation reports.
	Property string
	// Message is the message the constraint declares.
	Message string
	// Check reports whether the value satisfies the constraint. A nil value
	// satisfies every constraint here, matching JSR-380: @Max and friends
	// consider null valid, and only @NotNull rejects it.
	Check func(value any) bool
}

// Constrained is implemented by a model whose fields the original annotated
// with JSR-380 constraints. Go has no annotations, so the model states its own
// constraints.
type Constrained interface {
	Constraints() []Constraint
}

// JSR380Validator mirrors javax.validation.Validator: it evaluates the
// constraints a model declares.
type JSR380Validator struct{}

// BuildDefaultValidatorFactory mirrors
// javax.validation.Validation.buildDefaultValidatorFactory().GetValidator().
func BuildDefaultValidatorFactory() *JSR380Validator { return &JSR380Validator{} }

// Validate returns every violation of the target's declared constraints,
// ordered by property path so the result is stable — javax.validation returns
// a Set, whose iteration order is unspecified, and the application prints it.
func (v *JSR380Validator) Validate(target any) []ConstraintViolation {
	constrained, ok := target.(Constrained)
	if !ok {
		return nil
	}
	var violations []ConstraintViolation
	for _, c := range constrained.Constraints() {
		value := propertyValue(target, c.Property)
		if c.Check(value) {
			continue
		}
		violations = append(violations, ConstraintViolation{
			PropertyPath: c.Property,
			Message:      c.Message,
			InvalidValue: value,
		})
	}
	sort.SliceStable(violations, func(i, j int) bool {
		return violations[i].PropertyPath < violations[j].PropertyPath
	})
	return violations
}

// propertyValue reads a property through its getter, so a constraint names the
// property the way the bean exposes it.
func propertyValue(target any, property string) any {
	rv := reflect.ValueOf(target)
	getter := rv.MethodByName("Get" + capitalise(property))
	if !getter.IsValid() || getter.Type().NumIn() != 0 || getter.Type().NumOut() != 1 {
		return nil
	}
	out := getter.Call(nil)[0]
	if out.Kind() == reflect.Ptr && out.IsNil() {
		return nil
	}
	return out.Interface()
}

func capitalise(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
