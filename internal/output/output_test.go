package output

import (
	"bytes"
	"testing"
	"time"
)

func TestPrintTable(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		rows    [][]string
		want    string
	}{
		{
			name:    "empty table",
			headers: []string{},
			rows:    [][]string{},
			want:    "",
		},
		{
			name:    "headers only",
			headers: []string{"Name", "Age", "City"},
			rows:    [][]string{},
			want:    "Name   Age   City\n",
		},
		{
			name:    "single row",
			headers: []string{"Name", "Age"},
			rows:    [][]string{{"Alice", "30"}},
			want:    "Name    Age\nAlice   30\n",
		},
		{
			name:    "multiple rows",
			headers: []string{"Name", "Age"},
			rows: [][]string{
				{"Alice", "30"},
				{"Bob", "25"},
				{"Charlie", "35"},
			},
			want: "Name      Age\nAlice     30\nBob       25\nCharlie   35\n",
		},
		{
			name:    "no headers with rows",
			headers: []string{},
			rows: [][]string{
				{"Alice", "30"},
				{"Bob", "25"},
			},
			want: "Alice   30\nBob     25\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintTable(&buf, tt.headers, tt.rows)
			if err != nil {
				t.Fatalf("PrintTable() error = %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("PrintTable() output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestLocalTime(t *testing.T) {
	// Pin timezone so expected strings are deterministic across machines.
	t.Setenv("TZ", "Europe/Madrid")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "RFC3339 converts UTC to local timezone",
			input: "2025-06-15T10:30:00Z",
			want:  "2025-06-15 12:30:00 CEST",
		},
		{
			name:  "RFC3339 with offset",
			input: "2025-06-15T10:30:00+03:00",
			want:  "2025-06-15 09:30:00 CEST",
		},
		{
			name:  "RFC3339Nano",
			input: "2025-06-15T10:30:00.123456789Z",
			want:  "2025-06-15 12:30:00 CEST",
		},
		{
			name:  "microsecond precision format",
			input: "2025-06-15T10:30:00.123456",
			want:  "2025-06-15 12:30:00 CEST",
		},
		{
			name:  "bare datetime without fractional seconds",
			input: "2025-06-15T10:30:00",
			want:  "2025-06-15 12:30:00 CEST",
		},
		{
			name:  "returns original string on parse failure",
			input: "not-a-timestamp",
			want:  "not-a-timestamp",
		},
		{
			name:  "returns empty string unchanged",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LocalTime(tt.input); got != tt.want {
				t.Errorf("LocalTime(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoadingDots(t *testing.T) {
	tests := []struct {
		name  string
		ticks int
		want  string
	}{
		{name: "stop before any ticks", want: ""},
		{name: "stop waits for the first write", ticks: 1, want: "."},
		{name: "one dot per tick", ticks: 3, want: "..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			ticks := make(chan time.Time)
			stop := startLoadingDots(&buf, ticks)
			for range tt.ticks {
				ticks <- time.Time{}
			}
			stop()
			if got := buf.String(); got != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
