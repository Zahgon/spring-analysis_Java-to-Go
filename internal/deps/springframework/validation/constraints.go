package validation

import "fmt"

// Max builds the javax.validation.constraints.@Max constraint: the property
// must not exceed limit. A null value passes, as JSR-380 specifies.
func Max(property string, limit int64, message string) Constraint {
	return Constraint{
		Property: property,
		Message:  message,
		Check: func(value any) bool {
			n, ok := asInt64(value)
			if !ok {
				return true
			}
			return n <= limit
		},
	}
}

// asInt64 widens any integer value, reporting false for a null one.
func asInt64(value any) (int64, bool) {
	switch n := value.(type) {
	case nil:
		return 0, false
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case *int:
		if n == nil {
			return 0, false
		}
		return int64(*n), true
	default:
		panic(fmt.Sprintf("validation: unsupported numeric type %T", value))
	}
}
