package platform

import (
	"fmt"
	"strings"
	"testing"
)

func TestNotFound(t *testing.T) {
	many := make([]string, 25)
	for i := range many {
		many[i] = fmt.Sprintf("p%02d", i)
	}
	tests := []struct {
		names []string
		want  string
	}{
		{[]string{"zeta", "agents"}, `project "agent-email" not found; existing: agents, zeta`},
		{
			many,
			`project "agent-email" not found; existing: ` + strings.Join(many[:20], ", ") + ", …",
		},
	}
	for _, tt := range tests {
		if got := notFound("project", "agent-email", tt.names).Error(); got != tt.want {
			t.Errorf("got  %q\nwant %q", got, tt.want)
		}
	}
}
