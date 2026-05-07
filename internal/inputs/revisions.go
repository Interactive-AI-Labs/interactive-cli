package inputs

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseRevisionArg(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	rev, err := strconv.Atoi(raw)
	if err != nil || rev <= 0 {
		return 0, fmt.Errorf("revision must be a positive integer, got %q", raw)
	}
	return rev, nil
}
