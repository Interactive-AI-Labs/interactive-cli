package auth

import "fmt"

// KeyManagementLoginRequiredError reports that key management requires iai login.
func KeyManagementLoginRequiredError() error {
	return fmt.Errorf(
		"this command requires iai login; API key authentication is not supported for key management",
	)
}
