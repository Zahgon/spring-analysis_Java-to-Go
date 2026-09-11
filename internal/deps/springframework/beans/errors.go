package beans

import "fmt"

// NoSuchBeanDefinitionError is org.springframework.beans.factory.
// NoSuchBeanDefinitionException. Its message text is reproduced verbatim
// because it is what the base bootstrap demonstration prints before exiting.
type NoSuchBeanDefinitionError struct {
	// BeanName is set when the lookup was by name.
	BeanName string
	// TypeName is set when the lookup was by type.
	TypeName string
}

func (e *NoSuchBeanDefinitionError) Error() string {
	if e.TypeName != "" {
		return fmt.Sprintf("No qualifying bean of type '%s' available", e.TypeName)
	}
	return fmt.Sprintf("No bean named '%s' available", e.BeanName)
}

// NoUniqueBeanDefinitionError is thrown when a lookup by type matches more
// than one definition.
type NoUniqueBeanDefinitionError struct {
	TypeName string
	Names    []string
}

func (e *NoUniqueBeanDefinitionError) Error() string {
	return fmt.Sprintf("No qualifying bean of type '%s' available: expected single matching bean but found %d: %v",
		e.TypeName, len(e.Names), e.Names)
}

// BeanCreationError wraps a failure raised while a bean was being created.
type BeanCreationError struct {
	BeanName string
	Err      error
}

func (e *BeanCreationError) Error() string {
	return fmt.Sprintf("Error creating bean with name '%s': %v", e.BeanName, e.Err)
}

func (e *BeanCreationError) Unwrap() error { return e.Err }
