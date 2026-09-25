package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func printEndpoint(out io.Writer, endpoint *deployment.Endpoint) {
	if endpoint == nil || (endpoint.Private == "" && endpoint.Public == "") {
		return
	}
	fmt.Fprintln(out, "Endpoint:")
	if endpoint.Private != "" {
		fmt.Fprintf(out, "  Private:\t%s\n", endpoint.Private)
	}
	if endpoint.Public != "" {
		fmt.Fprintf(out, "  Public:\t%s\n", endpoint.Public)
	}
}
