package deployment

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAgentSessionCreateRequestJSON(t *testing.T) {
	tests := []struct {
		name  string
		input AgentSessionCreateRequest
		want  string
	}{
		{name: "guest", want: `{}`},
		{
			name: "customer and session fields",
			input: AgentSessionCreateRequest{
				CustomerID: "c1",
				Title:      "Order",
				Labels:     []string{"vip"},
			},
			want: `{"customerId":"c1","title":"Order","labels":["vip"]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("body = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDecodeAgentSession(t *testing.T) {
	const body = `{"id":"s1","customerId":"c1","createdAt":"2026-10-29T14:03:49Z","title":null,"mode":"auto","labels":["vip"],"metadata":{"number":9007199254740993}}`
	tests := []struct {
		name, body string
		want       *AgentSession
		wantErr    bool
	}{
		{
			name: "camelCase fields and exact raw metadata",
			body: body,
			want: &AgentSession{
				ID:         "s1",
				CustomerID: "c1",
				CreatedAt:  "2026-10-29T14:03:49Z",
				Mode:       "auto",
				Labels:     []string{"vip"},
				Raw:        json.RawMessage(body),
			},
		},
		{name: "invalid response", body: `[]`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeAgentSession([]byte(tt.body))
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, want error %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("session (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDecodeAgentSessionList(t *testing.T) {
	const body = `{"items":[{"id":"s1","customerId":"c1","createdAt":"2026-10-29T14:03:49Z"}],"totalCount":2,"hasMore":true,"nextCursor":"+/="}`
	const empty = `{"items":[],"totalCount":0,"hasMore":false,"nextCursor":null}`
	cursor := "+/="
	tests := []struct {
		name, body string
		want       *AgentSessionList
		wantErr    bool
	}{
		{
			name: "camelCase pagination and sessions", body: body,
			want: &AgentSessionList{
				Items: []AgentSession{
					{ID: "s1", CustomerID: "c1", CreatedAt: "2026-10-29T14:03:49Z"},
				},
				TotalCount: 2,
				HasMore:    true,
				NextCursor: &cursor,
				Raw:        json.RawMessage(body),
			},
		},
		{
			name: "empty page",
			body: empty,
			want: &AgentSessionList{Items: []AgentSession{}, Raw: json.RawMessage(empty)},
		},
		{name: "invalid response", body: `[`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeAgentSessionList([]byte(tt.body))
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, want error %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("list (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDecodeAgentEvent(t *testing.T) {
	const body = `{"id":"e1","source":"agent","kind":"message","offset":4,"createdAt":"2026-10-29T14:03:49Z","traceId":"t1","data":{"message":"Hello","status":"streaming","participant":{"displayName":"Hotel"}},"metadata":{"n":9007199254740993}}`
	tests := []struct {
		name, body string
		want       *AgentEvent
		wantErr    bool
	}{
		{
			name: "public event fields and exact raw payload", body: body,
			want: &AgentEvent{
				ID:        "e1",
				Source:    "agent",
				Kind:      "message",
				Offset:    4,
				CreatedAt: "2026-10-29T14:03:49Z",
				TraceID:   "t1",
				Data: json.RawMessage(
					`{"message":"Hello","status":"streaming","participant":{"displayName":"Hotel"}}`,
				),
				Raw: json.RawMessage(body),
			},
		},
		{name: "invalid response", body: `[`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeAgentEvent([]byte(tt.body))
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, want error %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("event (-want +got):\n%s", diff)
			}
		})
	}
}
