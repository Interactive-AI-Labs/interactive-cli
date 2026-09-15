package platform

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMcpEndpointWireShape(t *testing.T) {
	for _, tt := range []struct {
		name    string
		enabled bool
		want    string
	}{
		{"private", false, `{"image":"tools:1","port":8080,"endpoint":false,"path":"/mcp","memory":"128M","cpu":"100m"}`},
		{"public", true, `{"image":"tools:1","port":8080,"endpoint":true,"path":"/mcp","memory":"128M","cpu":"100m"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(
				McpWorkload{
					Image:    "tools:1",
					Port:     8080,
					Endpoint: tt.enabled,
					Path:     "/mcp",
					Memory:   "128M",
					CPU:      "100m",
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, string(got)); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestMcpEndpointResponse(t *testing.T) {
	internal := "http://tools.project.svc.cluster.local:8080/mcp"
	public := "tools.example.com"
	for _, tt := range []struct {
		name, body string
		want       *string
	}{
		{"public", `{"endpoint_url":"` + internal + `","endpoint":"` + public + `"}`, &public},
		{"private", `{"endpoint_url":"` + internal + `","endpoint":null}`, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got McpSchema
			if err := json.Unmarshal([]byte(tt.body), &got); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(
				McpSchema{EndpointURL: &internal, Endpoint: tt.want},
				got,
			); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
