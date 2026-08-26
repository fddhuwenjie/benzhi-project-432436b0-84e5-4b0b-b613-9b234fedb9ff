package domain

import "fmt"

type RuleError struct {
	Rule  string
	Cause error
}

func (e RuleError) Error() string { return fmt.Sprintf("%s: %v", e.Rule, e.Cause) }
func (e RuleError) Unwrap() error { return e.Cause }
