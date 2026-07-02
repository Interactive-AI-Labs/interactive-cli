package inputs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseVector(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    []float64
		wantErr bool
	}{
		{"simple", "0.1,0.2,0.3", []float64{0.1, 0.2, 0.3}, false},
		{"spaces", " 1 , 2 , 3 ", []float64{1, 2, 3}, false},
		{"trailing-comma", "1,2,", []float64{1, 2}, false},
		{"empty", "", nil, true},
		{"non-numeric", "1,a,3", nil, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseVector(c.in)
			if c.wantErr != (err != nil) {
				t.Fatalf("ParseVector(%q) err=%v, wantErr=%v", c.in, err, c.wantErr)
			}
			if !c.wantErr && !reflect.DeepEqual(got, c.want) {
				t.Errorf("ParseVector(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

func TestBuildBulkDeleteBody(t *testing.T) {
	cases := []struct {
		name    string
		ids     []string
		filter  string
		all     bool
		wantErr bool
		wantKey string // top-level key expected on success
	}{
		{"ids", []string{"a", "b"}, "", false, false, "ids"},
		{"filter", nil, `{"lang":"en"}`, false, false, "filter"},
		{"all", nil, "", true, false, "all"},
		{"none", nil, "", false, true, ""},
		{"two-selectors", []string{"a"}, "", true, true, ""},
		{"bad-filter", nil, `{not json`, false, true, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body, err := BuildBulkDeleteBody(c.ids, c.filter, c.all)
			if c.wantErr != (err != nil) {
				t.Fatalf("err=%v, wantErr=%v", err, c.wantErr)
			}
			if c.wantErr {
				return
			}
			var m map[string]any
			if uErr := json.Unmarshal(body, &m); uErr != nil {
				t.Fatalf("unmarshal: %v", uErr)
			}
			if _, ok := m[c.wantKey]; !ok {
				t.Errorf("body %s missing key %q", body, c.wantKey)
			}
		})
	}
}

func TestBuildSearchBody(t *testing.T) {
	cases := []struct {
		name    string
		query   string
		vector  []float64
		wantErr bool
	}{
		{"query-only", "hello", nil, false},
		{"vector-only", "", []float64{0.1, 0.2}, false},
		{"both", "hello", []float64{0.1}, true},
		{"neither", "", nil, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := BuildSearchBody(c.query, c.vector, "", 0, "")
			if c.wantErr != (err != nil) {
				t.Errorf("err=%v, wantErr=%v", err, c.wantErr)
			}
		})
	}
}

func TestBuildChunkCountBody(t *testing.T) {
	if _, err := BuildChunkCountBody("", ""); err != nil {
		t.Errorf("empty count body should not error: %v", err)
	}
	if _, err := BuildChunkCountBody(`{"lang":"en"}`, "doc-"); err != nil {
		t.Errorf("valid count body should not error: %v", err)
	}
	if _, err := BuildChunkCountBody(`{bad`, ""); err == nil {
		t.Errorf("bad filter should error")
	}
}

func TestBuildAddSlotBody(t *testing.T) {
	if _, err := BuildAddSlotBody("float32", 1536, "cosine"); err != nil {
		t.Errorf("valid slot body should not error: %v", err)
	}
	if _, err := BuildAddSlotBody("float32", 0, ""); err == nil {
		t.Errorf("dimension 0 should error")
	}
}

func TestReadCollectionBodyFile(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
		wantErr string
	}{
		{
			name:    "yaml config",
			content: "full_text:\n  enabled: true\n",
			want:    `{"full_text":{"enabled":true}}`,
		},
		{
			name:    "numeric metadata key becomes a string",
			content: "metadata:\n  2024: budget\n",
			want:    `{"metadata":{"2024":"budget"}}`,
		},
		{name: "empty file rejected", content: "", wantErr: "has no content"},
		{name: "comment-only file rejected", content: "# notes\n", wantErr: "has no content"},
		{name: "malformed rejected", content: "a: [b", wantErr: "failed to parse"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "body.yaml")
			if err := os.WriteFile(path, []byte(c.content), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := ReadCollectionBodyFile(path)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if string(got) != c.want {
				t.Errorf("body = %s, want %s", got, c.want)
			}
		})
	}
}
