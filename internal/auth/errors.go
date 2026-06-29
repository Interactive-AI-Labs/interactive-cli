package auth

import "fmt"

// KeyManagementLoginRequiredError reports that key management requires user auth.
func KeyManagementLoginRequiredError() error {
	return fmt.Errorf(
		"this command requires iai login or JWT authentication; API key authentication is not supported for key management",
	)
}
