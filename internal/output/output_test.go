package output

import (
	"bytes"
	"strings"
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
			name:  "converts UTC to local timezone in RFC1123Z format",
			input: "2025-06-15T10:30:00Z",
			want:  "Sun, 15 Jun 2025 12:30:00 +0200",
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

func TestPrintLoadingDots(t *testing.T) {
	t.Run("prints dots periodically", func(t *testing.T) {
		var buf bytes.Buffer
		done := PrintLoadingDots(&buf)

		time.Sleep(2500 * time.Millisecond)
		close(done)

		time.Sleep(100 * time.Millisecond)

		output := buf.String()
		dotCount := strings.Count(output, ".")
		if dotCount < 2 {
			t.Errorf("expected at least 2 dots, got %d. Output: %q", dotCount, output)
		}
		if dotCount > 3 {
			t.Errorf("expected at most 3 dots, got %d. Output: %q", dotCount, output)
		}
	})

	t.Run("stops when done channel is closed", func(t *testing.T) {
		var buf bytes.Buffer
		done := PrintLoadingDots(&buf)

		time.Sleep(500 * time.Millisecond)
		close(done)

		time.Sleep(500 * time.Millisecond)
		lenBefore := buf.Len()

		time.Sleep(1500 * time.Millisecond)
		lenAfter := buf.Len()

		if lenAfter != lenBefore {
			t.Errorf(
				"expected no more output after done, but buffer grew from %d to %d bytes",
				lenBefore,
				lenAfter,
			)
		}
	})
}
