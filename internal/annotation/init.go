// Package annotation holds the application's own annotation, ported to the
// construct Go offers in its place.
package annotation

// Init marks a method as a custom initialisation callback.
//
// The original is a @Retention(RUNTIME) @Target(METHOD) annotation, read back
// off the class by reflection. Go has no annotations and no way to attach
// metadata to a method, so a type that wants the same marking declares it:
// implementing Initializer is the Go equivalent of writing @Init above a
// method.
type Initializer interface {
	// InitMethods returns the names of this type's @Init-marked methods.
	InitMethods() []string
}

// Name is how the marker was spelled in the original, kept so that anything
// reporting on it prints the same text.
const Name = "@Init"

// MethodsOf returns the @Init-marked methods of v, or nil when it declares
// none — the answer reflection gave for a class with no annotated method.
func MethodsOf(v any) []string {
	marked, ok := v.(Initializer)
	if !ok {
		return nil
	}
	return marked.InitMethods()
}
