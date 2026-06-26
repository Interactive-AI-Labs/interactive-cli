package output

import "testing"

func TestFormatUSD(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{nil, ""},
		{"100", "$100.00"},
		{"100.00", "$100.00"},
		{1.5, "$1.50"},
		{"n/a", "n/a"},
	}

	for _, tt := range cases {
		if got := formatUSD(tt.in); got != tt.want {
			t.Fatalf("formatUSD(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
