package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// confirmDeletion tolerates io.EOF so input without a trailing newline (echo -n y) still counts.
func confirmDeletion(in io.Reader, out io.Writer, target string) (bool, error) {
	// No terminal means nobody can answer: an idle pipe would block forever.
	if f, ok := in.(*os.File); ok && !term.IsTerminal(int(f.Fd())) {
		return false, fmt.Errorf(
			"cannot ask whether to delete %s: stdin is not a terminal; re-run with -f to confirm",
			target,
		)
	}
	fmt.Fprintf(out, "This will delete %s. Continue? [y/N] ", target)
	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	return strings.ToLower(strings.TrimSpace(answer)) == "y", nil
}
