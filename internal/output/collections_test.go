package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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
			if got := humanBytes(c.in); got != c.want {
				t.Errorf("humanBytes(%d) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"shorter-than-n", "hello", 10, "hello"},
		{"exact-n", "hello", 5, "hello"},
		{"rune-cut", "hello world", 5, "hello…"},
		{"multi-byte-safe", "héllo wörld", 5, "héllo…"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := truncate(c.in, c.n); got != c.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", c.in, c.n, got, c.want)
			}
		})
	}
}

func TestPrintCollectionList(t *testing.T) {
	cases := []struct {
		name        string
		collections []clients.CollectionSummary
		wantLines   []string
	}{
		{
			name:        "empty",
			collections: nil,
			wantLines:   []string{"No collections found."},
		},
		{
			name: "two rows",
			collections: []clients.CollectionSummary{
				{
					Name:      "alpha",
					CreatedAt: "2026-01-01T00:00:00Z",
					UpdatedAt: "2026-01-02T00:00:00Z",
				},
				{
					Name:      "beta",
					CreatedAt: "2026-02-01T00:00:00Z",
					UpdatedAt: "2026-02-02T00:00:00Z",
				},
			},
			// Header plus one line per row, in declaration order. LocalTime
			// formatting is timezone-dependent so we only check structure.
			wantLines: []string{"NAME", "alpha", "beta"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintCollectionList(&buf, c.collections); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			for _, want := range c.wantLines {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\n%s", want, got)
				}
			}
		})
	}
}

func TestPrintCollectionStats(t *testing.T) {
	cases := []struct {
		name     string
		stats    *clients.CollectionStats
		wantSubs []string
	}{
		{
			name:     "no-index-valid",
			stats:    &clients.CollectionStats{ChunkCount: 0, SizeBytes: 0},
			wantSubs: []string{"Chunks:  0", "Size:    0 B"},
		},
		{
			name: "with-index-valid",
			stats: &clients.CollectionStats{
				ChunkCount: 42,
				SizeBytes:  5 * 1024 * 1024,
				IndexValid: map[string]bool{"default": true, "extra": false},
			},
			wantSubs: []string{
				"Chunks:  42",
				"Size:    5.0 MiB",
				"default",
				"true",
				"extra",
				"false",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintCollectionStats(&buf, c.stats); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			for _, want := range c.wantSubs {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\n%s", want, got)
				}
			}
		})
	}
}

func TestPrintCollectionDescribe(t *testing.T) {
	cases := []struct {
		name     string
		in       *clients.DescribeCollectionResponse
		wantSubs []string
	}{
		{
			name: "full-text-disabled",
			in: &clients.DescribeCollectionResponse{
				Name: "c",
				Config: clients.CollectionConfig{
					Vectors: map[string]clients.CollectionSlot{
						"default": {Type: "float32", Dimension: 4, Distance: "l2"},
					},
				},
			},
			wantSubs: []string{"Name:       c", "Full-text:  disabled", "float32", "deferred", "-"},
		},
		{
			name: "embedding-slot-with-index",
			in: &clients.DescribeCollectionResponse{
				Name: "c",
				Config: clients.CollectionConfig{
					Vectors: map[string]clients.CollectionSlot{
						"default": {
							Type:      "float32",
							Dimension: 1536,
							Distance:  "cosine",
							Embedding: &struct {
								Model string `json:"model"`
							}{Model: "interactive/openai/text-embedding-3-small"},
							Index: &clients.CollectionIndex{
								Type: "hnsw", M: 16, EfConstruction: 64,
							},
						},
					},
					FullText: &clients.CollectionFullText{Enabled: true, Language: "english"},
				},
			},
			wantSubs: []string{
				"Full-text:  enabled (english)",
				"interactive/openai/text-embedding-3-small",
				"hnsw (m=16, ef_construction=64)",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintCollectionDescribe(&buf, c.in); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			for _, want := range c.wantSubs {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\n%s", want, got)
				}
			}
		})
	}
}

func TestPrintChunkUpsertResult(t *testing.T) {
	cases := []struct {
		name string
		in   *clients.ChunkUpsertResult
		want string
	}{
		{
			name: "single-status",
			in: &clients.ChunkUpsertResult{
				Results: []clients.ChunkResult{
					{ID: "a", Status: "upserted"},
					{ID: "b", Status: "upserted"},
					{ID: "c", Status: "upserted"},
				},
			},
			want: "3 upserted\n",
		},
		{
			name: "no-results",
			in:   &clients.ChunkUpsertResult{},
			want: "",
		},
		{
			name: "multi-status-sorted",
			in: &clients.ChunkUpsertResult{
				Results: []clients.ChunkResult{
					{ID: "a", Status: "upserted"},
					{ID: "b", Status: "failed"},
					{ID: "c", Status: "upserted"},
					{ID: "d", Status: "created"},
				},
			},
			want: "1 created\n1 failed\n2 upserted\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintChunkUpsertResult(&buf, c.in); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != c.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, c.want)
			}
		})
	}
}

func TestPrintChunkList(t *testing.T) {
	cases := []struct {
		name     string
		in       *clients.ChunkList
		wantSubs []string
	}{
		{
			name:     "empty",
			in:       &clients.ChunkList{},
			wantSubs: []string{"No chunks found."},
		},
		{
			name: "single-page",
			in: &clients.ChunkList{
				Chunks: []clients.Chunk{
					{ID: "a", DocumentID: "doc", Text: "hello world"},
				},
			},
			wantSubs: []string{"ID", "DOCUMENT", "TEXT", "hello world"},
		},
		{
			name: "with-next-cursor",
			in: &clients.ChunkList{
				Chunks:     []clients.Chunk{{ID: "a", DocumentID: "doc", Text: "hello"}},
				HasMore:    true,
				NextCursor: stringPtr("opaque"),
			},
			wantSubs: []string{"hello", "More results", "opaque"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintChunkList(&buf, c.in); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			for _, want := range c.wantSubs {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\n%s", want, got)
				}
			}
		})
	}
}

func TestPrintChunk(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintChunk(&buf, &clients.Chunk{
		ID: "a", DocumentID: "doc",
		Text:     "hello",
		Metadata: map[string]any{"lang": "en"},
		Vector:   []float64{0.1, 0.2, 0.3},
		Vectors: map[string]json.RawMessage{
			"c-binary": json.RawMessage(`"0101"`),
			"b-sparse": json.RawMessage(`{"indices":[1,5],"values":[0.5,0.6],"dim":100}`),
			"a-dense":  json.RawMessage(`[0.1,0.2]`),
		},
	}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		"ID:        a\n",
		"Document:  doc\n",
		"Text:      hello\n",
		"\"lang\": \"en\"",
		// slots sorted by name; every vector kind summarized, none dropped
		"Vector[a-dense]: 2 dims\nVector[b-sparse]: sparse, 2 of 100 dims set\n" +
			"Vector[c-binary]: binary, 4 bits\n",
		"Vector:    3 dims\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\n%s", want, got)
		}
	}
}

func TestPrintBulkDeleteResult(t *testing.T) {
	cases := []struct {
		name string
		in   *clients.BulkDeleteResult
		want string
	}{
		{
			name: "count-only",
			in:   &clients.BulkDeleteResult{DeletedCount: 3},
			want: "Deleted 3 chunk(s)\n",
		},
		{
			name: "with-ids",
			in:   &clients.BulkDeleteResult{DeletedCount: 2, DeletedIds: []string{"a", "b"}},
			want: "Deleted 2 chunk(s)\n  a\n  b\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintBulkDeleteResult(&buf, c.in); err != nil {
				t.Fatal(err)
			}
			if got := buf.String(); got != c.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, c.want)
			}
		})
	}
}

func TestPrintDocumentList(t *testing.T) {
	cases := []struct {
		name     string
		in       *clients.DocumentList
		wantSubs []string
	}{
		{
			name:     "empty",
			in:       &clients.DocumentList{},
			wantSubs: []string{"No documents found."},
		},
		{
			name: "with-cursor",
			in: &clients.DocumentList{
				Documents:  []clients.DocumentSummary{{DocumentID: "doc", ChunkCount: 2}},
				HasMore:    true,
				NextCursor: stringPtr("next"),
			},
			wantSubs: []string{"DOCUMENT", "CHUNKS", "doc", "2", "More results", "next"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintDocumentList(&buf, c.in); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			for _, want := range c.wantSubs {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\n%s", want, got)
				}
			}
		})
	}
}

func TestPrintDocumentChunks(t *testing.T) {
	cases := []struct {
		name     string
		in       *clients.DocumentChunks
		wantSubs []string
	}{
		{
			name:     "empty",
			in:       &clients.DocumentChunks{DocumentID: "doc"},
			wantSubs: []string{"Document:  doc", "No chunks found."},
		},
		{
			name: "with-chunks",
			in: &clients.DocumentChunks{
				DocumentID: "doc",
				Chunks:     []clients.Chunk{{ID: "a", Text: "hello"}},
			},
			wantSubs: []string{"Document:  doc", "ID", "TEXT", "hello"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintDocumentChunks(&buf, c.in); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			for _, want := range c.wantSubs {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\n%s", want, got)
				}
			}
		})
	}
}

func TestPrintSearchResults(t *testing.T) {
	cases := []struct {
		name     string
		in       *clients.SearchResponse
		wantSubs []string
	}{
		{
			name:     "empty",
			in:       &clients.SearchResponse{},
			wantSubs: []string{"No results."},
		},
		{
			name: "two-hits",
			in: &clients.SearchResponse{
				Results: []clients.SearchHit{
					{ID: "a", Score: 0.9, Text: "first"},
					{ID: "b", Score: 0.5, Text: "second"},
				},
			},
			wantSubs: []string{"SCORE", "ID", "TEXT", "0.9000", "first", "0.5000", "second"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintSearchResults(&buf, c.in); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			for _, want := range c.wantSubs {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\n%s", want, got)
				}
			}
		})
	}
}

func TestPrintBatchSearchResults(t *testing.T) {
	cases := []struct {
		name     string
		in       *clients.BatchSearchResponse
		wantSubs []string
	}{
		{
			name:     "empty",
			in:       &clients.BatchSearchResponse{},
			wantSubs: []string{"No results."},
		},
		{
			name: "two-queries",
			in: &clients.BatchSearchResponse{
				Responses: []clients.SearchResponse{
					{Results: []clients.SearchHit{{ID: "a", Score: 0.9, Text: "first"}}},
					{Results: []clients.SearchHit{{ID: "b", Score: 0.5, Text: "second"}}},
				},
			},
			wantSubs: []string{"Query 1:", "Query 2:", "0.9000", "first", "0.5000", "second"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintBatchSearchResults(&buf, c.in); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			for _, want := range c.wantSubs {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\n%s", want, got)
				}
			}
		})
	}
}

func TestPrintSlotResults(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		var buf bytes.Buffer
		if err := PrintSlotAddResult(&buf, &clients.SlotAddResult{
			Slot: "s", Type: "float32", Dimension: 4, Distance: "l2", IndexStatus: "ready",
		}); err != nil {
			t.Fatal(err)
		}
		want := "Slot:          s\nType:          float32\nDimension:     4\nDistance:      l2\nIndex status:  ready\n"
		if got := buf.String(); got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("progress", func(t *testing.T) {
		var buf bytes.Buffer
		if err := PrintSlotIndexProgress(&buf, &clients.SlotIndexProgress{
			Slot: "s", IndexType: "hnsw", Status: "ready",
		}); err != nil {
			t.Fatal(err)
		}
		want := "Slot:        s\nIndex type:  hnsw\nStatus:      ready\n"
		if got := buf.String(); got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("op-with-status", func(t *testing.T) {
		var buf bytes.Buffer
		if err := PrintSlotOpResult(
			&buf,
			&clients.SlotOpResult{Slot: "s", Status: "ok"},
		); err != nil {
			t.Fatal(err)
		}
		if got := strings.TrimSpace(buf.String()); got != `Slot "s": ok` {
			t.Errorf("got %q want %q", got, `Slot "s": ok`)
		}
	})

	t.Run("op-with-index-status", func(t *testing.T) {
		var buf bytes.Buffer
		if err := PrintSlotOpResult(
			&buf,
			&clients.SlotOpResult{Slot: "s", IndexStatus: "ready"},
		); err != nil {
			t.Fatal(err)
		}
		if got := strings.TrimSpace(buf.String()); got != `Slot "s": ready` {
			t.Errorf("got %q want %q", got, `Slot "s": ready`)
		}
	})
}

func stringPtr(s string) *string { return &s }
