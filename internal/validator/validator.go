package validator

import (
	"regexp"
)

var (
	// Email address regular expression from https://html.spec.whatwg.org/#valid-e-mail-address
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Validator holds multiple errors for validating values in order to enforce business rules.
type Validator struct {
	Errors map[string]string
}

// New is a factory function returning a new Validator.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid checks if there are any errors recorded in the Validator.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError records an error key and message to the Validator.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check evaluates a boolean expression, recording an error in Validator if it evaluates to false.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Matches checks if a given string conforms to a regular expression.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
