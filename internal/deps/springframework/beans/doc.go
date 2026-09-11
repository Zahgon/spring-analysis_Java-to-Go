// Package beans reproduces the part of org.springframework.beans and
// org.springframework.context the application observes: a bean-definition
// registry with Spring's naming and ordering, lookup by name and by type
// (including its NoSuchBeanDefinitionException message), singleton and
// prototype scopes plus custom ones, and the BeanPostProcessor,
// BeanFactoryPostProcessor and Scope extension points.
//
// One thing Go cannot borrow from the JVM is Class.forName: there is no
// runtime that can turn the string "aop.SimpleAopBean" out of config.xml into
// a type. Classes therefore register themselves (see Register), which is the
// same information the JVM read out of the class files, written down instead
// of discovered.
package beans
