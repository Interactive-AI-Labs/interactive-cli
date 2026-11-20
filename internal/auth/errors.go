package auth

import "fmt"

// LoginRequiredError reports that an action requires user auth, not a project API key.
func LoginRequiredError(action string) error {
	return fmt.Errorf(
		"this command requires iai login or JWT authentication; API key authentication is not supported for %s",
		action,
	)
}
