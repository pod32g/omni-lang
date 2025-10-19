package cbackend

import "fmt"

// Print proxies to the runtime printing primitive. It currently only reports
// that the runtime has not been linked yet.
func Print(s string) error {
	if s == "" {
		return fmt.Errorf("print requires non-empty input")
	}
	return fmt.Errorf("c runtime backend: not implemented")
}
