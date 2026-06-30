package output

import "testing"

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		name string
		in   int64
		want string
	}{
		{"zero", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"exact-kib", 1024, "1.0 KiB"},
		{"kib", 204800, "200.0 KiB"},
		{"mib", 5 * 1024 * 1024, "5.0 MiB"},
		{"gib", 3 * 1024 * 1024 * 1024, "3.0 GiB"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := humanBytes(c.in); got != c.want {
				t.Errorf("humanBytes(%d) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
