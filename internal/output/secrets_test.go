package output

import (
	"bytes"
	"encoding/base64"
	"testing"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintSecretList(t *testing.T) {
	tests := []struct {
		name    string
		secrets []clients.SecretInfo
		want    string
	}{
		{
			name:    "empty list prints message",
			secrets: []clients.SecretInfo{},
			want:    "No secrets found.\n",
		},
		{
			name:    "nil list prints message",
			secrets: nil,
			want:    "No secrets found.\n",
		},
		{
			name: "single secret with few keys",
			secrets: []clients.SecretInfo{
				{
					Name:      "db-creds",
					Type:      "Opaque",
					CreatedAt: "2024-01-01",
					Keys:      []string{"USER", "PASS"},
				},
			},
			want: "NAME       TYPE     CREATED      KEYS\n" +
				"db-creds   Opaque   2024-01-01   USER, PASS\n",
		},
		{
			name: "secret with many keys truncates",
			secrets: []clients.SecretInfo{
				{
					Name:      "big-secret",
					Type:      "Opaque",
					CreatedAt: "2024-06-01",
					Keys:      []string{"A", "B", "C", "D", "E"},
				},
			},
			want: "NAME         TYPE     CREATED      KEYS\n" +
				"big-secret   Opaque   2024-06-01   A, B, C (+2 more)\n",
		},
		{
			name: "secret with exactly 3 keys shows all",
			secrets: []clients.SecretInfo{
				{
					Name:      "three-keys",
					Type:      "Opaque",
					CreatedAt: "2024-06-01",
					Keys:      []string{"X", "Y", "Z"},
				},
			},
			want: "NAME         TYPE     CREATED      KEYS\n" +
				"three-keys   Opaque   2024-06-01   X, Y, Z\n",
		},
		{
			name: "secret with no keys shows empty",
			secrets: []clients.SecretInfo{
				{
					Name:      "empty-secret",
					Type:      "Opaque",
					CreatedAt: "2024-06-01",
					Keys:      []string{},
				},
			},
			want: "NAME           TYPE     CREATED      KEYS\n" +
				"empty-secret   Opaque   2024-06-01   \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintSecretList(&buf, tt.secrets)
			if err != nil {
				t.Fatalf("PrintSecretList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintSecretData(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "empty data prints message",
			data: map[string]string{},
			want: "No data found in secret.\n",
		},
		{
			name: "nil data prints message",
			data: nil,
			want: "No data found in secret.\n",
		},
		{
			name: "base64 encoded values are decoded",
			data: map[string]string{
				"API_KEY": base64.StdEncoding.EncodeToString([]byte("my-secret-key")),
			},
			want: "KEYS      VALUES\n" +
				"API_KEY   my-secret-key\n",
		},
		{
			name: "non-base64 values are shown as-is",
			data: map[string]string{
				"PLAIN": "not-base64!!!",
			},
			want: "KEYS    VALUES\n" +
				"PLAIN   not-base64!!!\n",
		},
		{
			name: "keys are sorted alphabetically",
			data: map[string]string{
				"ZEBRA": "z",
				"ALPHA": "a",
				"MANGO": "m",
			},
			want: "KEYS    VALUES\n" +
				"ALPHA   a\n" +
				"MANGO   m\n" +
				"ZEBRA   z\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintSecretData(&buf, tt.data)
			if err != nil {
				t.Fatalf("PrintSecretData() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestFormatSecretKeys(t *testing.T) {
	tests := []struct {
		name       string
		keys       []string
		maxVisible int
		want       string
	}{
		{
			name:       "empty keys",
			keys:       []string{},
			maxVisible: 3,
			want:       "",
		},
		{
			name:       "nil keys",
			keys:       nil,
			maxVisible: 3,
			want:       "",
		},
		{
			name:       "fewer than max",
			keys:       []string{"A", "B"},
			maxVisible: 3,
			want:       "A, B",
		},
		{
			name:       "exactly max",
			keys:       []string{"A", "B", "C"},
			maxVisible: 3,
			want:       "A, B, C",
		},
		{
			name:       "more than max",
			keys:       []string{"A", "B", "C", "D", "E"},
			maxVisible: 3,
			want:       "A, B, C (+2 more)",
		},
		{
			name:       "one over max",
			keys:       []string{"X", "Y", "Z", "W"},
			maxVisible: 3,
			want:       "X, Y, Z (+1 more)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSecretKeys(tt.keys, tt.maxVisible)
			if got != tt.want {
				t.Errorf("formatSecretKeys() = %q, want %q", got, tt.want)
			}
		})
	}
}
