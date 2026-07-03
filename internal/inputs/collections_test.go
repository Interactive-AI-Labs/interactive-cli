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
	cases := []struct {
		name    string
		filter  string
		prefix  string
		want    string
		wantErr bool
	}{
		{name: "empty counts all", want: "{}"},
		{
			name:   "filter and prefix",
			filter: `{"lang":"en"}`,
			prefix: "doc-",
			want:   `{"filter":{"lang":"en"},"prefix":"doc-"}`,
		},
		{name: "prefix only", prefix: "doc-", want: `{"prefix":"doc-"}`},
		{name: "bad filter", filter: `{bad`, wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := BuildChunkCountBody(c.filter, c.prefix)
			if c.wantErr {
				if err == nil {
					t.Fatalf("BuildChunkCountBody(%q, %q) expected error", c.filter, c.prefix)
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

func TestBuildAddSlotBody(t *testing.T) {
	cases := []struct {
		name      string
		slotType  string
		dimension int
		distance  string
		want      string
		wantErr   bool
	}{
		{
			name:     "type, dimension, distance",
			slotType: "float32", dimension: 1536, distance: "cosine",
			want: `{"dimension":1536,"distance":"cosine","type":"float32"}`,
		},
		{
			name:     "distance omitted",
			slotType: "float32", dimension: 1536,
			want: `{"dimension":1536,"type":"float32"}`,
		},
		{name: "zero dimension errors", slotType: "float32", dimension: 0, wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := BuildAddSlotBody(c.slotType, c.dimension, c.distance)
			if c.wantErr {
				if err == nil {
					t.Fatalf("BuildAddSlotBody(%q, %d, %q) expected error",
						c.slotType, c.dimension, c.distance)
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
