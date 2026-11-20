package output

import (
	"fmt"
	"testing"
)

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
		{"just-under-mib-rounds-up-a-unit", 1024*1024 - 1, "1.0 MiB"},
		{"just-under-gib-rounds-up-a-unit", 1024*1024*1024 - 1, "1.0 GiB"},
		{"under-round-threshold-stays", 1048524, "1023.9 KiB"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := HumanBytes(c.in); got != c.want {
				t.Errorf("HumanBytes(%d) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestFormatUSD(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{nil, ""},
		{"100", "$100.00"},
		{"100.00", "$100.00"},
		{1.5, "$1.50"},
		{0.0375, "$0.0375"},
		{"n/a", "n/a"},
		{true, "true"},
	}

	for _, tt := range cases {
		if got := formatUSD(tt.in); got != tt.want {
			t.Fatalf("formatUSD(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatTokenCount(t *testing.T) {
	tests := []struct {
		in   *int
		want string
	}{
		{nil, "-"},
		{intPtr(0), "0"},
		{intPtr(999), "999"},
		{intPtr(1000), "1k"},
		{intPtr(32768), "33k"},
		{intPtr(64000), "64k"},
		{intPtr(999499), "999k"},
		{intPtr(999500), "1.0M"},
		{intPtr(1048576), "1.0M"},
		{intPtr(1050000), "1.1M"},
		{intPtr(1150000), "1.2M"},
		{intPtr(12_340_000), "12.3M"},
	}

	for _, tt := range tests {
		name := "nil"
		if tt.in != nil {
			name = fmt.Sprint(*tt.in)
		}
		t.Run(name, func(t *testing.T) {
			if got := formatTokenCount(tt.in); got != tt.want {
				t.Errorf("formatTokenCount(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
